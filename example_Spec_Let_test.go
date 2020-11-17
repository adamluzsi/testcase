package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Let() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Let(`variable Name`, func(t *testcase.T) interface{} {
		return "value that needs complex construction or can be mutated"
	})

	s.Then(`test case`, func(t *testcase.T) {
		t.Log(t.I(`variable Name`).(string)) // -> "value"
	})
}

func ExampleSpec_Let_eagerLoading() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Let(`variable Name`, func(t *testcase.T) interface{} {
		return "value that will be eager loaded before the test/then block reached"
	}).EagerLoading(s)

	s.Then(`test case`, func(t *testcase.T) {
		t.Log(t.I(`variable Name`).(string))
	})
}
