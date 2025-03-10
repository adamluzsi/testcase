package let

import "go.llib.dev/testcase"

// ElementFrom
//
// Deprecated: use let.OneOf instead
func ElementFrom[V any](s *testcase.Spec, vs ...V) testcase.Var[V] {
	return OneOf(s, vs...)
}
