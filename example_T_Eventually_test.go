package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/fixtures"
)

func ExampleT_Eventually() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	s.Test(``, func(t *testcase.T) {
		// Eventually this will pass eventually
		t.Eventually(func(it assert.It) {
			it.Must.True(fixtures.Random.Bool())
		})
	})
}
