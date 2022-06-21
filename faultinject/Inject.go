package faultinject

import (
	"context"
	"fmt"
	"reflect"
)

const (
	panicTagIsNil         = "Nil faultinject.Tag type received"
	panicTagNotStructType = "Invalid faultinject.Tag type received, got %T, but expected struct type"
)

// Inject will arrange context to trigger fault injection for the provided tags.
func Inject(ctx context.Context, tags ...Tag) context.Context {
	if !Enabled() {
		return ctx
	}
	if len(tags) == 0 {
		return ctx
	}
	for _, tag := range tags {
		if tag == nil {
			panic(panicTagIsNil)
		}
		if reflect.TypeOf(tag).Kind() != reflect.Struct {
			panic(fmt.Sprintf(panicTagNotStructType, tag))
		}
	}
	if v, ok := lookup(ctx); ok {
		*v = append(*v, tags...)
		return ctx
	}
	return context.WithValue(ctx, ctxKeyTags{}, &tags)
}

type ctxKeyTags struct{}

func lookup(ctx context.Context) (*[]Tag, bool) {
	if ctx == nil {
		return nil, false
	}
	v, ok := ctx.Value(ctxKeyTags{}).(*[]Tag)
	return v, ok
}
