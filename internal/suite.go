package internal

import (
	"testing"
)

type SuiteNullTB struct{ testing.TB }

func (n SuiteNullTB) Helper() {}

func (n SuiteNullTB) Cleanup(f func()) {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Error(args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Errorf(format string, args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Fail() {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) FailNow() {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Failed() bool {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Fatal(args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Fatalf(format string, args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Log(args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Logf(format string, args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Name() string {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Setenv(key, value string) {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Skip(args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) SkipNow() {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Skipf(format string, args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) Skipped() bool {
	//TODO implement me
	panic("implement me")
}

func (n SuiteNullTB) TempDir() string {
	//TODO implement me
	panic("implement me")
}
