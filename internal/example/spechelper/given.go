package spechelper

import "go.llib.dev/testcase"

func GivenWeHaveSomething(s *testcase.Spec) testcase.Var[any] {
	return testcase.Let(s, func(t *testcase.T) interface{} {
		// use user manager to create random user with fixtures maybe
		return nil
	})
}
