package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Describe(t *testing.T) {
	s := testcase.NewSpec(t)

	myType := func(v *testcase.V) *MyType {
		return &MyType{Field1: v.I(`input`).(string)}
	}

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(v *testcase.V) bool { return myType(v).IsLower() }

		s.When(`input string has lower case characters`, func(s *testcase.Spec) {
			s.Let(`input`, func(v *testcase.V) interface{} { return `all lower case` })

			s.Then(`it will return true`, func(t *testing.T, v *testcase.V) {
				t.Parallel()

				if subject(v) != true {
					t.Fatalf(`it was expected that the %q will re reported to be lowercase`, v.I(`input`))
				}
			})

			s.And(`the first character is capitalized`, func(s *testcase.Spec) {
				s.Let(`input`, func(v *testcase.V) interface{} { return `First character is uppercase` })

				s.Then(`it will report false`, func(t *testing.T, v *testcase.V) {
					if subject(v) != false {
						t.Fatalf(`it was expected that %q will be reported to be not lowercase`, v.I(`input`))
					}
				})
			})
		})
	})
}
