package doubles

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"go.llib.dev/testcase/sandbox"

	"go.llib.dev/testcase/internal/env"

	"go.llib.dev/testcase/internal/teardown"
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
	OnSkipNow    func()

	ctx       context.Context
	ctxCancel func()

	td    teardown.Teardown
	mutex sync.Mutex

	Tests []*TB

	passes int
}

func (m *TB) init() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.ctx == nil {
		m.ctx, m.ctxCancel = context.WithCancel(context.Background())
	}
}

func (m *TB) Finish() {
	m.init()
	m.ctxCancel()
	m.td.Finish()
}

func (m *TB) Context() context.Context {
	m.init()
	return m.ctx
}

func (m *TB) Cleanup(f func()) { m.td.Defer(f) }

// Chdir calls os.Chdir(dir) and uses Cleanup to restore the current
// working directory to its original value after the test.
// It also sets PWD environment variable for the duration of the test.
func (m *TB) Chdir(dir string) {
	m.Helper()

	og, err := os.Getwd()
	if err != nil {
		m.Fatal(err.Error())
		return
	}
	m.Cleanup(func() {
		if err := os.Chdir(og); err != nil {
			m.Fatal(err.Error())
			return
		}
	})

	if err := os.Chdir(dir); err != nil {
		m.Fatal(err.Error())
		return
	}

	current, err := os.Getwd()
	if err != nil {
		m.Fatal(err.Error())
		return
	}
	env.SetEnv(m, "PWD", current)
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
	if m.OnSkipNow != nil {
		m.OnSkipNow()
	}
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
		name = fmt.Sprintf("%d", len(m.Tests))
	}
	dtb := &TB{TB: m.TB, StubName: m.Name() + "/" + name}
	m.Tests = append(m.Tests, dtb)
	sandbox.Run(func() { blk(dtb) })
	if dtb.IsFailed {
		m.Error(dtb.Logs.String())
	}
	return !dtb.IsFailed
}

func (m *TB) LastRunTB() (*TB, bool) {
	if len(m.Tests) == 0 {
		return nil, false
	}
	return m.Tests[len(m.Tests)-1], true
}

func (m *TB) LastTB() *TB {
	if ltb, ok := m.LastRunTB(); ok {
		return ltb
	}
	return m
}

func (m *TB) Pass() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.passes++
}

func (m *TB) Passes() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.passes
}
