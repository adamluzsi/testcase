package faultinject

import (
	"context"
	"fmt"
	"sync"

	"go.llib.dev/testcase/internal/reflects"
)

// Inject will arrange context to trigger fault injection for the provided fault.
func Inject(ctx context.Context, fault any, err error) context.Context {
	if !Enabled() {
		return ctx
	}
	injectCTX, ok := lookupInjectContext(ctx)
	if !ok {
		injectCTX = newInjectContext(ctx)
		ctx = injectCTX
	}
	injectCTX.addTag(fault, err)
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

func newInjectContext(ctx context.Context) *injectContext {
	ctx, cancel := context.WithCancel(ctx)
	return &injectContext{Context: ctx, cancel: cancel}
}

type injectContext struct {
	context.Context
	mutex  sync.RWMutex
	faults faultCases
	err    error
	cancel func()
}

type faultCases map[any]error

const (
	panicFaultIsNil           = "Nil fault type is received"
	panicFaultIsNotStructType = "Invalid fault type is received, got %T, but expected struct type"
)

func (c *injectContext) Done() <-chan struct{} {
	_, _ = c.check()
	return c.Context.Done()
}

func (c *injectContext) Err() error {
	if err, ok := c.check(); ok {
		return err
	}
	return c.Context.Err()
}

func (c *injectContext) Value(key any) any {
	if key == (ctxKeyInjectContext{}) {
		return c
	}
	if _, ok := key.(*int); ok { // prevent cancelCtx lookup hack to bypass fault injection.
		if _, ok := c.Context.Value(key).(context.Context); ok {
			return nil
		}
	}
	_, _ = c.check(key)
	if err, ok := c.fetchBy(c.filterByFaults(key)); ok {
		return err
	}
	return c.Context.Value(key)
}

func (c *injectContext) check(faults ...any) (error, bool) {
	if err := c.getError(); err != nil {
		return err, true
	}
	if err, ok := c.checkForCallerFaults(); ok {
		c.setError(err)
		return err, ok
	}
	if err, ok := c.fetchBy(c.filterByFaults(faults...)); ok {
		c.setError(err)
		return err, ok
	}
	return nil, false
}

func (c *injectContext) getError() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.err
}

func (c *injectContext) setError(err error) {
	c.mutex.Lock()
	c.err = err
	c.mutex.Unlock()
	c.cancel()
	<-c.Context.Done()
	//
	// it is nearly impossible to know when will the context.cancelCtx signal itself that this context is cancelled,
	// so the least wrong I was able to come up without coupling or hacking is to schedule the go routines a couple of times.
	// If someone uses Context.Done to wait before checking the error, this should be a safe operation.
	wait()
}

func (c *injectContext) addTag(fault any, err error) {
	if reflects.IsNil(fault) {
		panic(panicFaultIsNil)
	}
	if !reflects.IsStruct(fault) {
		panic(fmt.Sprintf(panicFaultIsNotStructType, fault))
	}
	if c.faults == nil {
		c.faults = make(faultCases)
	}
	if err == nil {
		err = DefaultErr
	}
	c.faults[fault] = err
}

func (c *injectContext) fetchBy(filter func(fault any) bool) (error, bool) {
	var (
		rErr error
		ok   bool
	)
	for fault, err := range c.faults {
		if has := filter(fault); !has {
			continue
		}
		rErr = err
		ok = true
		break
	}
	return rErr, ok
}

func (c *injectContext) filterByFaults(faults ...any) func(fault any) bool {
	return func(fault any) bool {
		for _, target := range faults {
			if fault == target {
				return true
			}
		}
		return false
	}
}

func (c *injectContext) checkForCallerFaults() (error, bool) {
	return c.fetchBy(func(tag any) bool {
		f, ok := tag.(CallerFault)
		if !ok {
			return false
		}
		return f.check()
	})
}
