package faultinject_test

import (
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/faultinject"
)

var enabled testcase.Var[bool]

func init() {
	enabled = testcase.Var[bool]{
		ID: "faultinject is enabled",
		Init: func(t *testcase.T) bool {
			return true
		},
		OnLet: func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				if enabled.Get(t) {
					faultinject.EnableForTest(t)
				}
			})
		},
	}
}
