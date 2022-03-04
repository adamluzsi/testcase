package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleAppend() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	list := testcase.Let(s, func(t *testcase.T) interface{} {
		return []int{}
	})

	s.Before(func(t *testcase.T) {
		t.Log(`some context where a value is expected in the testcase.Var[[]T] variable`)
		testcase.Append(t, list, 42)
	})

	s.Test(``, func(t *testcase.T) {
		t.Log(`list will include the appended value`)
		list.Get(t) // []int{42}
	})
}
