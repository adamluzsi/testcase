package testcase_test

import (
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
)

func ExampleGroup() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Context(`description`, func(s *testcase.Spec) {

		s.Test(``, func(t *testcase.T) {})

	}, testcase.Group(`testing-group-group-that-can-be-even-targeted-with-testCase-run-cli-option`))
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

func ExampleFlaky_retryUntilTimeout() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test(`testCase with "random" fails`, func(t *testcase.T) {
		// This testCase might fail "randomly" but the retry flag will allow some tolerance
		// This should be used to find time in team's calendar
		// and then allocate time outside of death-march times to learn to avoid retry tests in the future.
	}, testcase.Flaky(time.Minute))
}

func ExampleFlaky_retryNTimes() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test(`testCase with "random" fails`, func(t *testcase.T) {
		// This testCase might fail "randomly" but the retry flag will allow some tolerance
		// This should be used to find time in team's calendar
		// and then allocate time outside of death-march times to learn to avoid retry tests in the future.
	}, testcase.Flaky(42))
}
