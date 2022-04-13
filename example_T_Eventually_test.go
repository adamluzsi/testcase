package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
)

func ExampleT_Eventually() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	s.Test(``, func(t *testcase.T) {
		// Eventually this will pass eventually
		t.Eventually(func(it assert.It) {
			it.Must.True(t.Random.Bool())
		})
	})
}
