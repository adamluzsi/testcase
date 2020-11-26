package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_SkipBenchmark() {
	var b *testing.B
	s := testcase.NewSpec(b)
	s.SkipBenchmark()

	s.Test(`this will be skipped during benchmark`, func(t *testcase.T) {})

	s.Context(`some context`, func(s *testcase.Spec) {
		s.Test(`this as well`, func(t *testcase.T) {})
	})
}

func ExampleSpec_SkipBenchmark_scopedWithContext() {
	var b *testing.B
	s := testcase.NewSpec(b)

	s.When(`rainy path`, func(s *testcase.Spec) {
		s.SkipBenchmark()

		s.Test(`will be skipped during benchmark`, func(t *testcase.T) {})
	})

	s.Context(`happy path`, func(s *testcase.Spec) {
		s.Test(`this will run as benchmark`, func(t *testcase.T) {})
	})
}
