package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	// . "my/project/testing/pkg"
)

func ExampleSpec_whenProjectUseSharedSpecificationHelpers() {
	var t *testing.T
	s := testcase.NewSpec(t)
	SetupSpec(s)

	GivenWeHaveUser(s) // Order
	// .. other givens

	myType := func() *MyType { return &MyType{} }

	s.Describe(`#MyFunc`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) { myType().MyFunc() } // Act

		s.Then(`edge case description`, func(t *testcase.T) {
			// Assert
			subject(t)
		})
	})
}

/*
	------------------------------------------------------------------------
	Somewhere else in a project's testing package ("my/project/testing/pkg")
	------------------------------------------------------------------------
*/

func SetupSpec(s *testcase.Spec) {
	testcase.Let(s, func(t *testcase.T) interface{} {
		// create new storage connection
		// t.Defer(s.Close) after the storage was used in the testCase
		return nil
	})
	testcase.Let(s, func(t *testcase.T) interface{} {
		// new user manager with storage
		return nil
	})
}

func GivenWeHaveUser(s *testcase.Spec) testcase.Var[any] {
	return testcase.Let(s, func(t *testcase.T) interface{} {
		// use user manager to create random user with fixtures maybe
		return nil
	})
}
