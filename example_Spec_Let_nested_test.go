package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Let_usageWithinNestedScope() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var myType = func(t *testcase.T) *MyType { return &MyType{} }

	s.Describe(`#IsLower`, func(s *testcase.Spec) {
		var (
			input   = testcase.Var{Name: `input`}
			subject = func(t *testcase.T) bool {
				return myType(t).IsLower(input.Get(t).(string))
			}
		)

		s.When(`input characters are list lowercase`, func(s *testcase.Spec) {
			s.Let(`input`, func(t *testcase.T) interface{} {
				return "list lowercase"
			})
			// or
			input.Let(s, func(t *testcase.T) interface{} {
				return "list lowercase"
			})

			s.Then(`it will report true`, func(t *testcase.T) {
				t.Must.True(subject(t))
			})
		})

		s.When(`input is a capitalized`, func(s *testcase.Spec) {
			s.Let(`input`, func(t *testcase.T) interface{} {
				return "Capitalized"
			})
			// or
			input.Let(s, func(t *testcase.T) interface{} {
				return "Capitalized"
			})

			s.Then(`it will report false`, func(t *testcase.T) {
				t.Must.True(!subject(t))
			})
		})
	})
}
