package doubles

import (
	"bytes"
	"fmt"
	"github.com/adamluzsi/testcase/sandbox"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/adamluzsi/testcase/internal/env"

	"github.com/adamluzsi/testcase/internal/teardown"
)

type TB struct {
	// TB is an optional value here.
	// If provided, some default behaviour might be taken from it, like TempDir.
	//
	// It also helps implement testing.TB interface's with embedding.
	testing.TB

	IsFailed  bool
	IsSkipped bool
	Logs      bytes.Buffer

	StubName     string
	StubNameFunc func() string
	StubTempDir  string
	OnFailNow    func()

	td    teardown.Teardown
	mutex sync.Mutex

	RunTBs []*TB
}

func (m *TB) Finish() {
	m.td.Finish()
}

func (m *TB) Cleanup(f func()) {
	m.td.Defer(f)
}

func (m *TB) Error(args ...any) {
	m.appendLogs(fmt.Sprintln(args...))
	m.Fail()
}

func (m *TB) Errorf(format string, args ...any) {
	m.appendLogs(fmt.Sprintf(format+"\n", args...))
	m.Fail()
}

func (m *TB) Fail() {
	m.IsFailed = true
}

func (m *TB) FailNow() {
	if m.OnFailNow != nil {
		m.OnFailNow()
	}
	m.Fail()
	runtime.Goexit()
}

func (m *TB) Failed() bool {
	return m.IsFailed
}

func (m *TB) Fatal(args ...any) {
	m.appendLogs(fmt.Sprintln(args...))
	m.FailNow()
}

func (m *TB) Fatalf(format string, args ...any) {
	m.appendLogs(fmt.Sprintf(format+"\n", args...))
	m.FailNow()
}

func (m *TB) Helper() {}

func (m *TB) appendLogs(msg string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	_, _ = fmt.Fprint(&m.Logs, msg)
}

func (m *TB) Log(args ...any) {
	m.appendLogs(fmt.Sprintln(args...))
}

func (m *TB) Logf(format string, args ...any) {
	m.appendLogs(fmt.Sprintf(format+"\n", args...))
}

func (m *TB) Name() string {
	if m.StubNameFunc != nil {
		return m.StubNameFunc()
	}
	if m.StubName == "" {
		m.StubName = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return m.StubName
}

func (m *TB) Skip(args ...any) {
	m.Log(args...)
	m.SkipNow()
}

func (m *TB) SkipNow() {
	m.IsSkipped = true
	runtime.Goexit()
}

func (m *TB) Skipf(format string, args ...any) {
	m.Logf(format, args...)
	m.SkipNow()
}

func (m *TB) Skipped() bool {
	return m.IsSkipped
}

func (m *TB) TempDir() string {
	if m.StubTempDir != "" {
		return m.StubTempDir
	}
	if m.TB == nil {
		return os.TempDir()
	}
	return m.TB.TempDir()
}

func (m *TB) Setenv(key, value string) {
	env.SetEnv(m, key, value)
}

func (m *TB) Run(name string, blk func(tb testing.TB)) bool {
	if name == "" {
		name = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	dtb := &TB{TB: m.TB, StubName: m.Name() + "/" + name}
	m.RunTBs = append(m.RunTBs, dtb)
	sandbox.Run(func() { blk(dtb) })
	if dtb.IsFailed {
		m.Error(dtb.Logs.String())
	}
	return !dtb.IsFailed
}

func (m *TB) LastRunTB() (*TB, bool) {
	if len(m.RunTBs) == 0 {
		return nil, false
	}
	return m.RunTBs[len(m.RunTBs)-1], true
}

func (m *TB) LastTB() *TB {
	if ltb, ok := m.LastRunTB(); ok {
		return ltb
	}
	return m
}
