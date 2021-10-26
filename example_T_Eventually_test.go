package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
)

func ExampleT_Eventually() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	s.Test(``, func(t *testcase.T) {
		// Eventually this will pass
		t.Eventually(func(tb testing.TB) {
			if fixtures.Random.Bool() {
				tb.FailNow()
			}
		})
	})
}
