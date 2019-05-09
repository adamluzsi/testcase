package testcase_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Let(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Let(`variable name`, func(t *testcase.T) interface{} {
		return "value"
	})

	s.Then(`test case`, func(t *testcase.T) {
		t.Log(t.I(`variable name`).(string)) // -> "value"
	})
}

func ExampleSpec_Let_usageWithinANestedConext(t *testing.T) {
	myType := func(t *testcase.T) *MyType { return &MyType{Field1: t.I(`input`).(string)} }

	s := testcase.NewSpec(t)

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) bool { return myType(t).IsLower() }

		s.When(`input characters are all lowercase`, func(s *testcase.Spec) {
			s.Let(`input`, func(t *testcase.T) interface{} {
				return "all lowercase"
			})

			s.Then(`it will report true`, func(t *testcase.T) {
				require.True(t, subject(t))
			})
		})

		s.When(`input is a capitalized`, func(s *testcase.Spec) {
			s.Let(`input`, func(t *testcase.T) interface{} {
				return "Capitalized"
			})

			s.Then(`it will report false`, func(t *testcase.T) {
				require.False(t, subject(t))
			})
		})
	})
}