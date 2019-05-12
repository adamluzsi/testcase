package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Let(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Let(`variable name`, func(t *testcase.T) interface{} {
		return "value"
	})

	s.Then(`test case`, func(t *testcase.T) {
		t.Log(t.I(`variable name`).(string)) // -> "value"
	})
}
