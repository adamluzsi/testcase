package faultinject

import (
	"context"
)

// Inject will arrange context to trigger fault injection for the provided tags.
func Inject(ctx context.Context, tags ...string) context.Context {
	var faults []fault
	for _, tag := range tags {
		faults = append(faults, fault{Tag: tag})
	}
	if v, ok := lookup(ctx); ok {
		*v = append(*v, faults...)
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, &faults)
}

type ctxKey struct{}

func lookup(ctx context.Context) (*[]fault, bool) {
	if ctx == nil {
		return nil, false
	}
	v, ok := ctx.Value(ctxKey{}).(*[]fault)
	return v, ok
}
