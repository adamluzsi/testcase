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

func ExampleSpec_BeforeAll() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.BeforeAll(func(tb testing.TB) {
		// this will run once before every test cases.
	})
}

func ExampleSpec_AfterAll() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.AfterAll(func(tb testing.TB) {
		// this will run once all the test case already ran.
	})
}

func ExampleSpec_AroundAll() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.AroundAll(func(tb testing.TB) func() {
		// this will run once before all the test case.
		return func() {
			// this will run once after all the test case already ran.
		}
	})
}
