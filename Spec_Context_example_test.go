package testcase_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Context(t *testing.T) {
	s := testcase.NewSpec(t)

	myType := func(t *testcase.T) *MyType {
		return &MyType{Field1: t.I(`input`).(string)}
	}

	s.Context(`describe IsLower`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) bool { return myType(t).IsLower() }

		s.Context(`when lowercase`, func(s *testcase.Spec) {
			s.Let(`input`, func(t *testcase.T) interface{} {
				return `lowercase text`
			})

			s.Then(`test-case`, func(t *testcase.T) {
				require.True(t, subject(t))
			})
		})
	})
}
