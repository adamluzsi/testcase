package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_And() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.When(`some spec`, func(s *testcase.Spec) {
		// fulfil the spec

		s.And(`additional spec`, func(s *testcase.Spec) {

			s.Then(`assert`, func(t *testcase.T) {

			})
		})

		s.And(`additional spec opposite`, func(s *testcase.Spec) {

			s.Then(`assert`, func(t *testcase.T) {

			})
		})
	})
}
