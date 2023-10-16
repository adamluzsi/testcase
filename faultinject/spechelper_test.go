package faultinject_test

import (
	"go.llib.dev/testcase"
	"go.llib.dev/testcase/faultinject"
)

var enabled = testcase.Var[bool]{
	ID: "faultinject is enabled",
	Init: func(t *testcase.T) bool {
		return true
	},
	OnLet: func(s *testcase.Spec, enabled testcase.Var[bool]) {
		s.Before(func(t *testcase.T) {
			if enabled.Get(t) {
				faultinject.EnableForTest(t)
			}
		})
	},
}

var exampleErr = testcase.Var[error]{
	ID: "example error",
	Init: func(t *testcase.T) error {
		return t.Random.Error()
	},
}
