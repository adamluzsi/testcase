package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Sequential() {
	var t *testing.T
	s := testcase.NewSpec(t)
	s.Sequential() // tells the specs to run list test case in sequence

	s.Test(`this will run in sequence`, func(t *testcase.T) {})

	s.Context(`some spec`, func(s *testcase.Spec) {
		s.Test(`this run in sequence`, func(t *testcase.T) {})

		s.Test(`this run in sequence`, func(t *testcase.T) {})
	})
}

func ExampleSpec_Sequential_scopedWithContext() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Parallel() // on top level, spec marked as parallel

	s.Context(`spec marked sequential`, func(s *testcase.Spec) {
		s.Sequential() // but in subcontext the testCase marked as sequential

		s.Test(`this run in sequence`, func(t *testcase.T) {})
	})

	s.Context(`spec that inherit parallel flag`, func(s *testcase.Spec) {

		s.Test(`this will run in parallel`, func(t *testcase.T) {})
	})
}

func ExampleSpec_HasSideEffect() {
	var t *testing.T
	s := testcase.NewSpec(t)
	// this mark the testCase to contain side effects.
	// this forbids any parallel testCase execution to avoid retry tests.
	//
	// Under the hood this is a syntax sugar for Sequential
	s.HasSideEffect()

	s.Test(`this will run in sequence`, func(t *testcase.T) {})

	s.Context(`some spec`, func(s *testcase.Spec) {
		s.Test(`this run in sequence`, func(t *testcase.T) {})

		s.Test(`this run in sequence`, func(t *testcase.T) {})
	})
}
