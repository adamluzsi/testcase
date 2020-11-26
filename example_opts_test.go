package testcase_test

import (
	"github.com/adamluzsi/testcase"
	"testing"
)

func ExampleName() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Context(`description`, func(s *testcase.Spec) {

		s.Test(``, func(t *testcase.T) {})

	}, testcase.Name(`name-that-can-be-targeted-with-test-run`))
}

func ExampleSkipBenchmark() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test(`will run`, func(t *testcase.T) {
		// this will run during benchmark execution
	})

	s.Test(`will skip`, func(t *testcase.T) {
		// this will skip the benchmark execution
	}, testcase.SkipBenchmark())
}
