package faultinject

import (
	"context"
	"fmt"

	"github.com/adamluzsi/testcase/internal/reflects"
)

// Inject will arrange context to trigger fault injection for the provided fault.
func Inject(ctx context.Context, fault any, err error) context.Context {
	if !Enabled() {
		return ctx
	}
	ictx, ok := lookupInjectContext(ctx)
	if !ok {
		ctx, ictx = withInjectContext(ctx)
	}
	ictx.addTag(fault, err)
	return ctx
}

type ctxKeyInjectContext struct{}

func lookupInjectContext(ctx context.Context) (*injectContext, bool) {
	if ctx == nil {
		return nil, false
	}
	v, ok := ctx.Value(ctxKeyInjectContext{}).(*injectContext)
	return v, ok
}

func withInjectContext(ctx context.Context) (context.Context, *injectContext) {
	ictx := &injectContext{Context: ctx}
	return ictx, ictx
}

type injectContext struct {
	context.Context
	faults faultCases
	err    error
}

type faultCases map[any]error

const (
	panicFaultIsNil           = "Nil fault type is received"
	panicFaultIsNotStructType = "Invalid fault type is received, got %T, but expected struct type"
)

func (ictx *injectContext) addTag(fault any, err error) {
	if reflects.IsNil(fault) {
		panic(panicFaultIsNil)
	}
	if !reflects.IsStruct(fault) {
		panic(fmt.Sprintf(panicFaultIsNotStructType, fault))
	}
	if ictx.faults == nil {
		ictx.faults = make(faultCases)
	}
	if err == nil {
		err = DefaultErr
	}
	ictx.faults[fault] = err
}

func (ictx *injectContext) fetchBy(filter func(fault any) bool) (error, bool) {
	var (
		rErr error
		ok   bool
	)
	for fault, err := range ictx.faults {
		if has := filter(fault); !has {
			continue
		}
		rErr = err
		ok = true
		break
	}
	return rErr, ok
}

func (ictx *injectContext) Done() <-chan struct{} {
	if ictx.Err() != nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return ictx.Context.Done()
}

func (ictx *injectContext) Err() error {
	if err := ictx.Context.Err(); err != nil {
		return err
	}
	if ictx.err == nil {
		if err, ok := ictx.checkForFaults(); ok {
			ictx.err = err
		}
	}
	return ictx.err
}

func (ictx *injectContext) checkForFaults() (error, bool) {
	return ictx.fetchBy(func(tag any) bool {
		f, ok := tag.(CallerFault)
		if !ok {
			return false
		}
		return f.check()
	})
}

func (ictx *injectContext) Value(key any) any {
	if key == (ctxKeyInjectContext{}) {
		return ictx
	}
	if err, ok := ictx.checkForFaults(); ok {
		return err
	}
	if err, ok := ictx.fetchBy(func(fault any) bool {
		return fault == key
	}); ok {
		return err
	}
	return ictx.Context.Value(key)
}
