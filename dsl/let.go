package dsl

import (
	"github.com/adamluzsi/testcase"
)

func Let[V any](spec *testcase.Spec, blk func(*testcase.T) V) testcase.Var[V] {
	return testcase.Let[V](spec, blk)
}

func LetValue[V any](spec *testcase.Spec, value V) testcase.Var[V] {
	return testcase.LetValue[V](spec, value)
}
