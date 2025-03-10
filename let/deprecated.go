package let

import "go.llib.dev/testcase"

// ElementFrom
//
// DEPRECATED: use let.OneOf instead
func ElementFrom[V any](s *testcase.Spec, vs ...V) testcase.Var[V] {
	return OneOf(s, vs...)
}
