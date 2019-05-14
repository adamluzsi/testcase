package testcase_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Run(t *testing.T) {
	s := testcase.NewSpec(t)

	myType := func(t *testcase.T) *MyType {
		return &MyType{Field1: `input`}
	}

	s.Context(`describe IsLower`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) bool { return myType(t).IsLower() }

		s.Context(`when something`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { /* setup */ })

			s.Then(`test-case`, func(t *testcase.T) {
				require.True(t, subject(t))
			})
		})
	})
}
