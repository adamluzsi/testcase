package testcase_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Context() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		myType  = func(t *testcase.T) *MyType { return &MyType{} }
		subject = func(t *testcase.T) bool { return myType(t).IsLower(t.I(`input`).(string)) }
	)

	s.Context(`when input is in lowercase`, func(s *testcase.Spec) {
		s.LetValue(`input`, `lowercase text`)

		s.Then(`test-case`, func(t *testcase.T) {
			require.True(t, subject(t))
		})
	})
}
