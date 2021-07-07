package internal

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

type StubTB struct {
	// TB is only present here to implement testing.TB interface's
	// unexported functions by embedding the interface itself.
	testing.TB

	IsFailed  bool
	IsSkipped bool
	Logs      []string

	StubName    string
	StubTempDir string
	StubFailNow func()

	td    Teardown
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
	if m.StubFailNow != nil {
		m.StubFailNow()
	} else {
		runtime.Goexit()
	}
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
	return m.StubTempDir
}
