// Package let contains Common Testcase Variable Let declarations for testing purpose.
package let

import (
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/internal"
	"github.com/adamluzsi/testcase/random"
)

func Contact(s *testcase.Spec, opts ...internal.ContactOption) testcase.Var[random.Contact] {
	return testcase.Let[random.Contact](s, func(t *testcase.T) random.Contact {
		return t.Random.Contact(opts...)
	})
}

func FirstName(s *testcase.Spec, opts ...internal.ContactOption) testcase.Var[string] {
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.Contact(opts...).FirstName
	})
}

func LastName(s *testcase.Spec) testcase.Var[string] {
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.Contact().LastName
	})
}

func Email(s *testcase.Spec) testcase.Var[string] {
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.Contact().Email
	})
}
