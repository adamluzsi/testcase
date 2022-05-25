package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleT_SkipUntil() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Test(`will be skipped`, func(t *testcase.T) {
		// make tests skip until the given day is reached,
		// then make the tests fail.
		// This helps to commit code which still work in progress.
		t.SkipUntil(2020, 01, 01)
	})

	s.Test(`will not be skipped`, func(t *testcase.T) {})
}
