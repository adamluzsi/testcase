package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_LetValue() {
	var t *testing.T
	s := testcase.NewSpec(t)

	variable := testcase.LetValue(s, "value")

	s.Then(`test case`, func(t *testcase.T) {
		t.Log(variable.Get(t)) // -> "value"
	})
}
