package testcase_test

import (
	"github.com/adamluzsi/testcase"
	"testing"
	"time"
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

func ExampleFlaky() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test(`test with "random" fails`, func(t *testcase.T) {
		// This test might fail "randomly" but the flaky flag will allow some tolerance
		// This should be used to find time in team's calendar
		// and then allocate time outside of death-march times to learn to avoid flaky tests in the future.
	}, testcase.Flaky(time.Minute))
}