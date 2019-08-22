package testcase_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Let_usageWithinNestedScope(t *testing.T) {
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
