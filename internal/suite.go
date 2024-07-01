package internal

import (
	"testing"
)

type NullTB struct{ testing.TB }

func (n NullTB) Helper() {}

func (n NullTB) Cleanup(f func()) {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Error(args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Errorf(format string, args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Fail() {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) FailNow() {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Failed() bool {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Fatal(args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Fatalf(format string, args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Log(args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Logf(format string, args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Name() string {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Setenv(key, value string) {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Skip(args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) SkipNow() {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Skipf(format string, args ...any) {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) Skipped() bool {
	//TODO implement me
	panic("implement me")
}

func (n NullTB) TempDir() string {
	//TODO implement me
	panic("implement me")
}
