package testcase_test

import (
	"github.com/adamluzsi/testcase"
	"testing"
	"time"
)

func ExampleGroup() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Context(`description`, func(s *testcase.Spec) {

		s.Test(``, func(t *testcase.T) {})

	}, testcase.Group(`testing-group-name-that-can-be-even-targeted-with-test-run-cli-option`))
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