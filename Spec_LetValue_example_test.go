package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_LetValue(t *testing.T) {
	s := testcase.NewSpec(t)

	s.LetValue(`variable name`, "value")

	s.Then(`test case`, func(t *testcase.T) {
		t.Log(t.I(`variable name`).(string)) // -> "value"
	})
}
