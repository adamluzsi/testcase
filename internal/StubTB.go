package internal

import (
	"runtime"
	"testing"
)

type StubTB struct {
	// TB is only present here to implement testing.TB interface's
	// unexported functions by embedding the interface itself.
	testing.TB

	IsFailed  bool
	IsSkipped bool

	StubName    string
	StubTempDir string

	td Teardown
}

func (m *StubTB) Finish() {
	m.td.Finish()
}

func (m *StubTB) Cleanup(f func()) {
	m.td.Cleanup(f)
}

func (m *StubTB) Error(args ...interface{}) {
	m.Fail()
}

func (m *StubTB) Errorf(format string, args ...interface{}) {
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
	m.FailNow()
}

func (m *StubTB) Fatalf(format string, args ...interface{}) {
	m.FailNow()
}

func (m *StubTB) Helper() {}

func (m *StubTB) Log(args ...interface{}) {}

func (m *StubTB) Logf(format string, args ...interface{}) {}

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
