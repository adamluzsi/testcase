package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Before() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Before(func(t *testcase.T) {
		// this will run before the testCase cases.
	})
}

func ExampleSpec_After() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.After(func(t *testcase.T) {
		// this will run after the testCase cases.
		// this hook applied to this scope and anything that is nested from here.
		// hooks can be stacked with each call.
	})
}

func ExampleSpec_Around() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Around(func(t *testcase.T) func() {
		// this will run before the testCase cases

		// this hook applied to this scope and anything that is nested from here.
		// hooks can be stacked with each call
		return func() {
			// The content of the returned func will be deferred to run after the testCase cases.
		}
	})
}
