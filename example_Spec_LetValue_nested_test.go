package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_LetValue_usageWithinNestedScope() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var myType = func(t *testcase.T) *MyType { return &MyType{} }

	s.Describe(`#IsLower`, func(s *testcase.Spec) {
		var (
			input   = testcase.Var[string]{ID: `input`}
			subject = func(t *testcase.T) bool {
				return myType(t).IsLower(input.Get(t))
			}
		)

		s.When(`input characters are list lowercase`, func(s *testcase.Spec) {
			testcase.LetValue(s, "list lowercase")
			// or
			input.LetValue(s, "list lowercase")

			s.Then(`it will report true`, func(t *testcase.T) {
				t.Must.True(subject(t))
			})
		})

		s.When(`input is a capitalized`, func(s *testcase.Spec) {
			testcase.LetValue(s, "Capitalized")
			// or
			input.LetValue(s, "Capitalized")

			s.Then(`it will report false`, func(t *testcase.T) {
				t.Must.True(!subject(t))
			})
		})
	})
}
