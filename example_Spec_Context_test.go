package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Context() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Context(`description of the testing spec`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			// prepare for the testing spec
		})

		s.Then(`assert expected outcome`, func(t *testcase.T) {

		})
	})
}
