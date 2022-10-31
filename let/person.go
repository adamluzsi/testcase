// Package let contains Common Testcase Variable Let declarations for testing purpose.
//
package let

import (
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/internal"
)

func FirstName(s *testcase.Spec, opts ...internal.PersonOption) testcase.Var[string] {
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.Name().First(opts...)
	})
}

func LastName(s *testcase.Spec) testcase.Var[string] {
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.Name().Last()
	})
}

func Email(s *testcase.Spec) testcase.Var[string] {
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.Email()
	})
}
