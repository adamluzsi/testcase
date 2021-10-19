package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleT_must() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	s.Test(``, func(t *testcase.T) {
		// failed test will stop with FailNow
		t.Must.Equal(1, 1, "must be equal")
	})
}

func ExampleT_should() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	s.Test(``, func(t *testcase.T) {
		// failed test will proceed, but mart the test failed
		t.Should.Equal(1, 1, "should be equal")
	})
}
