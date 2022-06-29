package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

type VoidLogger struct{}

func (e VoidLogger) Log(args ...any) {}

func (e VoidLogger) Error(args ...any) {}

var Logger = testcase.Var[testcase.Logger]{
	ID: "some id",
	Init: func(t *testcase.T) testcase.Logger {
		return VoidLogger{}
	},
}

func ExampleSpec_withLogger() {
	var tb testing.TB
	s := testcase.NewSpec(tb, testcase.WithLogger(Logger))

	s.Test("", func(t *testcase.T) {
		t.Log("This goes into the void...")
	})
}
