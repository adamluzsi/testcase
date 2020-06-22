package testcase_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Let_usageWithinNestedScope() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var myType = func(t *testcase.T) *MyType { return &MyType{} }

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) bool {
			return myType(t).IsLower(t.I(`input`).(string))
		}

		s.When(`input characters are all lowercase`, func(s *testcase.Spec) {
			s.LetValue(`input`, `all lowercase`)

			s.Then(`it will report true`, func(t *testcase.T) {
				require.True(t, subject(t))
			})
		})

		s.When(`input is a capitalized`, func(s *testcase.Spec) {
			s.LetValue(`input`, "Capitalized")

			s.Then(`it will report false`, func(t *testcase.T) {
				require.False(t, subject(t))
			})
		})
	})
}
