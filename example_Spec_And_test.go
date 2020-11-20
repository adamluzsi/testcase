package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_And() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.When(`some context`, func(s *testcase.Spec) {
		// fulfil the context

		s.And(`additional context`, func(s *testcase.Spec) {

			s.Then(`assert`, func(t *testcase.T) {

			})
		})

		s.And(`additional context opposite`, func(s *testcase.Spec) {

			s.Then(`assert`, func(t *testcase.T) {

			})
		})
	})
}
