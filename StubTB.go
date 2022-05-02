package testcase

import (
	"fmt"
	"github.com/adamluzsi/testcase/internal"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

type StubTB struct {
	// TB is an optional value here.
	// If provided, some default behaviour might be taken from it, like TempDir.
	//
	// It also helps implement testing.TB interface's with embedding.
	testing.TB

	IsFailed  bool
	IsSkipped bool
	Logs      []string

	StubName    string
	StubTempDir string

	td    internal.Teardown
	mutex sync.Mutex
}

func (m *StubTB) Finish() {
	m.td.Finish()
}

func (m *StubTB) Cleanup(f func()) {
	m.td.Defer(f)
}

func (m *StubTB) Error(args ...interface{}) {
	m.appendLogs(fmt.Sprint(args...))
	m.Fail()
}

func (m *StubTB) Errorf(format string, args ...interface{}) {
	m.appendLogs(fmt.Sprintf(format, args...))
	m.Fail()
}

func (m *StubTB) Fail() {
	m.IsFailed = true
}

func (m *StubTB) FailNow() {
	m.Fail()
	runtime.Goexit()
}

func (m *StubTB) Failed() bool {
	return m.IsFailed
}

func (m *StubTB) Fatal(args ...interface{}) {
	m.appendLogs(fmt.Sprint(args...))
	m.FailNow()
}

func (m *StubTB) Fatalf(format string, args ...interface{}) {
	m.appendLogs(fmt.Sprintf(format, args...))
	m.FailNow()
}

func (m *StubTB) Helper() {}

func (m *StubTB) appendLogs(msg string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Logs = append(m.Logs, msg)
}

func (m *StubTB) Log(args ...interface{}) {
	m.appendLogs(fmt.Sprint(args...))
}

func (m *StubTB) Logf(format string, args ...interface{}) {
	m.appendLogs(fmt.Sprintf(format, args...))
}

func (m *StubTB) Name() string {
	if m.StubName == "" {
		m.StubName = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return m.StubName
}

func (m *StubTB) Skip(args ...interface{}) {
	m.SkipNow()
}

func (m *StubTB) SkipNow() {
	m.IsSkipped = true
	runtime.Goexit()
}

func (m *StubTB) Skipf(format string, args ...interface{}) {
	m.SkipNow()
}

func (m *StubTB) Skipped() bool {
	return m.IsSkipped
}

func (m *StubTB) TempDir() string {
	if m.StubTempDir != "" {
		return m.StubTempDir
	}
	if m.TB == nil {
		return os.TempDir()
	}
	return m.TB.TempDir()
}
