package testcase_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Before(t *testing.T) {
	myType := func(input string) *MyType { return &MyType{Field1: input} }

	s := testcase.NewSpec(t)

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(input string) bool { return myType(input).IsLower() }

		s.Before(func(t *testing.T, v *testcase.V) {
			// this will run before the test cases.
		})

		s.Then(`it will report whether Field1 is lower or not`, func(t *testing.T, v *testcase.V) {
			require.True(t, subject(`all lower case character`))
			require.False(t, subject(`Capitalized`))
		})
	})
}

func ExampleSpec_After(t *testing.T) {
	myType := func(input string) *MyType { return &MyType{Field1: input} }

	s := testcase.NewSpec(t)

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(input string) bool { return myType(input).IsLower() }

		s.After(func(t *testing.T, v *testcase.V) {
			// this will run after the test cases.
			// this hook applied to this scope and anything that is nested from here.
			// hooks can be stacked with each call.
		})

		s.Then(`it will report whether Field1 is lower or not`, func(t *testing.T, v *testcase.V) {
			require.True(t, subject(`all lower case character`))
			require.False(t, subject(`Capitalized`))
		})
	})
}

func ExampleSpec_Around(t *testing.T) {
	myType := func(input string) *MyType { return &MyType{Field1: input} }

	s := testcase.NewSpec(t)

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(input string) bool { return myType(input).IsLower() }

		s.Around(func(t *testing.T, v *testcase.V) func() {
			// this will run before the test cases

			// this hook applied to this scope and anything that is nested from here.
			// hooks can be stacked with each call
			return func() {
				// The content of the returned func will be deferred to run after the test cases.
			}
		})

		s.Then(`it will report whether Field1 is lower or not`, func(t *testing.T, v *testcase.V) {
			require.True(t, subject(`all lower case character`))
			require.False(t, subject(`Capitalized`))
		})
	})
}
