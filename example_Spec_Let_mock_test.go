package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/internal"
)


func ExampleSpec_Let_testingDouble() {
	var t *testing.T
	s := testcase.NewSpec(t)

	stub := s.Let(`my testing double`, func(t *testcase.T) interface{} {
		stub := &internal.StubTB{}
		t.Defer(stub.Finish)
		return stub
	})

	s.When(`some scope where double should behave in a certain way`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			stub.Get(t).(*internal.StubTB).StubName= "my stubbed name"
		})

		s.Then(`double will be available in every test case and finishNow called afterwards`, func(t *testcase.T) {
			// ...
		})
	})
}

type InterfaceExample interface {
	Say() string
}
