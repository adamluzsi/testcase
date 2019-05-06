package testcase_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Let(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Let(`variable name`, func(v *testcase.V) interface{} {
		return "value"
	})

	s.Then(`test case`, func(t *testing.T, v *testcase.V) {
		t.Log(v.I(`variable name`).(string)) // -> "value"
	})
}

func ExampleSpec_Let_usageWithinANestedConext(t *testing.T) {
	myType := func(v *testcase.V) *MyType { return &MyType{Field1: v.I(`input`).(string)} }

	s := testcase.NewSpec(t)

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(v *testcase.V) bool { return myType(v).IsLower() }

		s.When(`input characters are all lowercase`, func(s *testcase.Spec) {
			s.Let(`input`, func(v *testcase.V) interface{} {
				return "all lowercase"
			})

			s.Then(`it will report true`, func(t *testing.T, v *testcase.V) {
				require.True(t, subject(v))
			})
		})

		s.When(`input is a capitalized`, func(s *testcase.Spec) {
			s.Let(`input`, func(v *testcase.V) interface{} {
				return "Capitalized"
			})

			s.Then(`it will report false`, func(t *testing.T, v *testcase.V) {
				require.False(t, subject(v))
			})
		})
	})
}