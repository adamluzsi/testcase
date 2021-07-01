package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleNewT() {
	variable := testcase.Var{Name: "variable", Init: func(t *testcase.T) interface{} {
		return t.Random.Int()
	}}

	// flat test case with test runtime variable caching
	var tb testing.TB
	t := testcase.NewT(tb, testcase.NewSpec(tb))
	value1 := variable.Get(t).(int)
	value2 := variable.Get(t).(int)
	t.Logf(`test case variable caching works even in flattened tests: v1 == v2 -> %v`, value1 == value2)
}
