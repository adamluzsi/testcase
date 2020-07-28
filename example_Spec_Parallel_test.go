package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Parallel() {
	var t *testing.T
	s := testcase.NewSpec(t)
	s.Parallel() // tells the specs to run all test case in parallel

	s.Test(`this will run in parallel`, func(t *testcase.T) {})

	s.Context(`some context`, func(s *testcase.Spec) {
		s.Test(`this run in parallel`, func(t *testcase.T) {})

		s.Test(`this run in parallel`, func(t *testcase.T) {})
	})
}

func ExampleSpec_Parallel_scopedWithContext() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Context(`context marked parallel`, func(s *testcase.Spec) {
		s.Parallel()

		s.Test(`this run in parallel`, func(t *testcase.T) {})
	})

	s.Context(`context without parallel`, func(s *testcase.Spec) {

		s.Test(`this will run in sequence`, func(t *testcase.T) {})
	})
}

func ExampleSpec_NoSideEffect() {
	var t *testing.T
	s := testcase.NewSpec(t)
	// this is an idiom to express that the subject in the tests here are not expected to have any side-effect.
	// this means they are safe to be executed in parallel.
	s.NoSideEffect()

	s.Test(`this will run in parallel`, func(t *testcase.T) {})

	s.Context(`some context`, func(s *testcase.Spec) {
		s.Test(`this run in parallel`, func(t *testcase.T) {})

		s.Test(`this run in parallel`, func(t *testcase.T) {})
	})
}