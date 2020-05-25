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

	GivenWeHaveUser(s, `myuser`) // Arrange
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
