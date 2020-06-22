package testcase_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_When() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		myType  = func(t *testcase.T) *MyType { return &MyType{} }
		subject = func(t *testcase.T) bool { return myType(t).IsLower(t.I(`input`).(string)) }
	)

	s.When(`input has upcase letter`, func(s *testcase.Spec) {
		s.LetValue(`input`, `UPPER`)

		s.Then(`it will be false`, func(t *testcase.T) {
			require.False(t, subject(t))
		})
	})

	s.When(`input is all lowercase letter`, func(s *testcase.Spec) {
		s.LetValue(`input`, `lower`)

		s.Then(`it will be true`, func(t *testcase.T) {
			require.True(t, subject(t))
		})
	})
}
