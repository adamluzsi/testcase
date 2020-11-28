package internal

import (
	"runtime"
	"testing"
)

type StubTB struct {
	testing.TB
	IsFailed bool

	cleanups    []func()
	CleanupFunc func(func())
}

func (m *StubTB) Finish() {
	for _, fn := range m.cleanups {
		defer fn()
	}
}

func (m *StubTB) Cleanup(f func()) {
	if m.CleanupFunc == nil {
		m.cleanups = append(m.cleanups, f)
	} else {
		m.CleanupFunc(f)
	}
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

func (m *StubTB) Helper() {
	panic("implement me")
}

func (m *StubTB) Log(args ...interface{}) {
	panic("implement me")
}

func (m *StubTB) Logf(format string, args ...interface{}) {
	panic("implement me")
}

func (m *StubTB) Name() string {
	panic("implement me")
}

func (m *StubTB) Skip(args ...interface{}) {
	panic("implement me")
}

func (m *StubTB) SkipNow() {
	panic("implement me")
}

func (m *StubTB) Skipf(format string, args ...interface{}) {
	panic("implement me")
}

func (m *StubTB) Skipped() bool {
	panic("implement me")
}

func (m *StubTB) TempDir() string {
	panic("implement me")
}
