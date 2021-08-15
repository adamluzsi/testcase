package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Let() {
	var t *testing.T
	s := testcase.NewSpec(t)

	myTestVar := s.Let(`variable Name`, func(t *testcase.T) interface{} {
		return "value that needs complex construction or can be mutated"
	})

	s.Then(`test case`, func(t *testcase.T) {
		t.Log(myTestVar.Get(t).(string)) // -> returns the value set in the current spec spec for MyTestVar
	})
}

func ExampleSpec_Let_eagerLoading() {
	var t *testing.T
	s := testcase.NewSpec(t)

	myTestVar := s.Let(`variable Name`, func(t *testcase.T) interface{} {
		return "value that will be eager loaded before the testCase/then block reached"
	}).EagerLoading(s)
	// EagerLoading will ensure that the value of this Spec Var will be evaluated during the preparation of the testCase.

	s.Then(`test case`, func(t *testcase.T) {
		t.Log(myTestVar.Get(t).(string))
	})
}
