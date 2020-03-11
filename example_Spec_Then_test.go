package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Then() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Then(`it is expected.... so this is the test description here`, func(t *testcase.T) {
		// ...
	})
}
