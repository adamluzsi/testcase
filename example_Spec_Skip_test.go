package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Skip() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Context(`sub spec`, func(s *testcase.Spec) {
		s.Skip(`WIP`)

		s.Test(`will be skipped`, func(t *testcase.T) {})

		s.Test(`will be skipped as well`, func(t *testcase.T) {})

		s.Context(`skipped as well just like the tests of the parent`, func(s *testcase.Spec) {
			s.Test(`will be skipped`, func(t *testcase.T) {})
		})
	})

	s.Test(`this will still run since it is not part of the scope where Spec#Skip was called`, func(t *testcase.T) {})
}
