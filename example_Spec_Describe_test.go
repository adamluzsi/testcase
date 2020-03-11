package testcase_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Describe() {
	var t *testing.T
	s := testcase.NewSpec(t)

	myType := func(_ *testcase.T) *MyType {
		return &MyType{Field1: `input`}
	}

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) bool { return myType(t).IsLower() }

		s.Then(`test-case`, func(t *testcase.T) {
			// it will panic since `input` is not actually set at this testing scope,
			// and the testing framework will warn us about this.
			require.True(t, subject(t))
		})
	})
}
