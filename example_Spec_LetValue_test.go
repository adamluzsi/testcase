package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_LetValue() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.LetValue(`variable Name`, "value")

	s.Then(`test case`, func(t *testcase.T) {
		t.Log(t.I(`variable Name`).(string)) // -> "value"
	})
}
