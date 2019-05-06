package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Before(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Before(func(t *testing.T, v *testcase.V) {
		// this will run before the test cases.
	})
}

func ExampleSpec_After(t *testing.T) {
	s := testcase.NewSpec(t)

	s.After(func(t *testing.T, v *testcase.V) {
		// this will run after the test cases.
		// this hook applied to this scope and anything that is nested from here.
		// hooks can be stacked with each call.
	})
}

func ExampleSpec_Around(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Around(func(t *testing.T, v *testcase.V) func() {
		// this will run before the test cases

		// this hook applied to this scope and anything that is nested from here.
		// hooks can be stacked with each call
		return func() {
			// The content of the returned func will be deferred to run after the test cases.
		}
	})
}
