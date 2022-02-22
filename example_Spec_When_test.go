package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_When() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		myType  = func(t *testcase.T) *MyType { return &MyType{} }
		input   = testcase.Var[string]{Name: `input`}
		subject = func(t *testcase.T) bool { return myType(t).IsLower(input.Get(t)) }
	)

	s.When(`input has only upcase letter`, func(s *testcase.Spec) {
		input.LetValue(s, "UPPER")

		s.Then(`it will be false`, func(t *testcase.T) {
			t.Must.True(!subject(t))
		})
	})

	s.When(`input has only lowercase letter`, func(s *testcase.Spec) {
		input.LetValue(s, "lower")

		s.Then(`it will be true`, func(t *testcase.T) {
			t.Must.True(subject(t))
		})
	})
}
