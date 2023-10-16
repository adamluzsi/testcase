package let

import (
	"testing"

	"go.llib.dev/testcase"
)

func With[V any, FN withFN[V]](s *testcase.Spec, fn FN) testcase.Var[V] {
	var init testcase.VarInit[V]
	switch fnv := any(fn).(type) {
	case func() V:
		init = func(t *testcase.T) V { return fnv() }
	case func(testing.TB) V:
		init = func(t *testcase.T) V { return fnv(t) }
	case func(*testcase.T) V:
		init = fnv
	}
	return testcase.Let(s, init)
}

type withFN[V any] interface {
	func() V |
		func(testing.TB) V |
		func(*testcase.T) V
}
