package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Let_testingDouble() {
	var t *testing.T
	s := testcase.NewSpec(t)

	stubTB := testcase.Let(s, func(t *testcase.T) *testcase.StubTB {
		stub := &testcase.StubTB{}
		t.Defer(stub.Finish)
		return stub
	})

	s.When(`some scope where double should behave in a certain way`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			stubTB.Get(t).StubName = "my stubbed name"
		})

		s.Then(`double will be available in every test case and finishNow called afterwards`, func(t *testcase.T) {
			// ...
		})
	})
}

type InterfaceExample interface {
	Say() string
}
