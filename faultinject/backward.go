package faultinject

import "context"

// Finish is an alias for After
//
// Deprecated: use After instead
func Finish(returnErr *error, ctx context.Context, faults ...any) {
	After(returnErr, ctx, faults...)
}
