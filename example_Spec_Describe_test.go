package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Describe() {
	var t *testing.T
	s := testcase.NewSpec(t)

	myType := s.Let(`myType`, func(t *testcase.T) interface{} {
		return &MyType{}
	})

	// Describe description points orderingOutput the subject of the tests
	s.Describe(`#IsLower`, func(s *testcase.Spec) {
		var (
			input   = testcase.Var{Name: `input`}
			subject = func(t *testcase.T) bool {
				// subject should represent what will be tested in the describe block
				return myType.Get(t).(*MyType).IsLower(input.Get(t).(string))
			}
		)

		s.Test(``, func(t *testcase.T) { subject(t) })
	})
}
