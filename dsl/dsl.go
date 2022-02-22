package dsl

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
)

func Spec(tb testing.TB, opts ...testcase.SpecOption) *testcase.Spec {
	return testcase.NewSpec(tb, opts...)
}

func Let[V any](spec *testcase.Spec, varName string, blk func(*testcase.T) V) testcase.Var[V] {
	return testcase.Let[V](spec, varName, blk)
}
func LetValue[V any](spec *testcase.Spec, varName string, value V) testcase.Var[V] {
	return testcase.LetValue[V](spec, varName, value)
}

func Must(tb testing.TB) assert.Asserter   { return assert.Must(tb) }
func Should(tb testing.TB) assert.Asserter { return assert.Should(tb) }
