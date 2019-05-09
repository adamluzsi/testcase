package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Describe(t *testing.T) {
	s := testcase.NewSpec(t)

	myType := func(t *testcase.T) *MyType {
		return &MyType{Field1: t.I(`input`).(string)}
	}

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) bool { return myType(t).IsLower() }

		s.When(`input string has lower case characters`, func(s *testcase.Spec) {
			s.Let(`input`, func(t *testcase.T) interface{} { return `all lower case` })

			s.Then(`it will return true`, func(t *testcase.T) {
				t.Parallel()

				if subject(t) != true {
					t.Fatalf(`it was expected that the %q will re reported to be lowercase`, t.I(`input`))
				}
			})

			s.And(`the first character is capitalized`, func(s *testcase.Spec) {
				s.Let(`input`, func(t *testcase.T) interface{} { return `First character is uppercase` })

				s.Then(`it will report false`, func(t *testcase.T) {
					if subject(t) != false {
						t.Fatalf(`it was expected that %q will be reported to be not lowercase`, t.I(`input`))
					}
				})
			})
		})
	})
}
