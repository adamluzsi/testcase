package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Test() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Test(`my test description`, func(t *testcase.T) {
		// ...
	})
}
