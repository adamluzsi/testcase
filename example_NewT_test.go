package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleNewT() {
	variable := testcase.Var[int]{ID: "variable", Init: func(t *testcase.T) int {
		return t.Random.Int()
	}}

	// flat test case with test runtime variable caching
	var tb testing.TB
	t := testcase.NewT(tb, testcase.NewSpec(tb))
	value1 := variable.Get(t)
	value2 := variable.Get(t)
	t.Logf(`test case variable caching works even in flattened tests: v1 == v2 -> %v`, value1 == value2)
}
