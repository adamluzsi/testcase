package testcase_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"

	// . "my/project/testing/pkg"
)

func ExampleSpec_whenProjectUseSharedSpecificationHelpers() {
	var t *testing.T
	s := testcase.NewSpec(t)
	SetupSpec(s)

	GivenWeHaveUser(s, `myuser`)
	// .. other givens

	s.Describe(`#Myfunc`, func(s *testcase.Spec) {
		// define describe's subject

		s.Then(`edge case description`, func(t *testcase.T) {
			t.I(`myuser`) // can be accessed here
		})
	})

	myType := func(_ *testcase.T) *MyType { return &MyType{} }

	s.Describe(`Something`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) bool { return myType(t).IsLower() }

		s.Then(`test-case`, func(t *testcase.T) {
			// it will panic since `input` is not actually set at this testing scope,
			// and the testing framework will warn us about this.
			require.True(t, subject(t))
		})
	})
}

/*
	------------------------------------------------------------------------
	Somewhere else in a project's testing package ("my/project/testing/pkg")
	------------------------------------------------------------------------
*/

func SetupSpec(s *testcase.Spec) {
	s.Let(`storage`, func(t *testcase.T) interface{} {
		// create new storage connection
		// t.Defer(s.Close) after the storage was used in the test
		return nil
	})
	s.Let(`user manager`, func(t *testcase.T) interface{} {
		// new user manager with storage
		return nil
	})
}

func GivenWeHaveUser(s *testcase.Spec, userLetVar string) {
	s.Let(userLetVar, func(t *testcase.T) interface{} {
		// use user manager to create random user with fixtures maybe
		return nil
	})
}
