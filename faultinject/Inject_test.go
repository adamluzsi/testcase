package faultinject_test

import (
	"context"
	"errors"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/faultinject"
	"github.com/adamluzsi/testcase/random"
	"testing"
)

func TestInject_smoke(t *testing.T) {
	t.Cleanup(faultinject.Enable())

	assert.NotPanic(t, func() {
		type ExampleTag struct{}
		faultinject.Inject(context.Background(), ExampleTag{}, errors.New("boom"))
	})
	assert.NotPanic(t, func() {
		type ExampleTag struct{ Error error }
		faultinject.Inject(context.Background(), ExampleTag{Error: nil}, errors.New("boom"))
	})
}

func TestInject_onEnabledFalse(t *testing.T) {
	assert.False(t, faultinject.Enabled())
	inCTX := context.Background()
	outCTX := faultinject.Inject(inCTX, FaultTag{}, errors.New("boom"))
	assert.Equal(t, inCTX, outCTX)
}

func TestInject_nilErrRetrunsDefaultErr(t *testing.T) {
	faultinject.EnableForTest(t)
	ctx := faultinject.Inject(context.Background(), FaultTag{}, nil)
	assert.Equal[error](t, faultinject.DefaultErr, ctx.Value(FaultTag{}).(error))
	ctx = faultinject.Inject(context.Background(), faultinject.CallerFault{}, nil)
	assert.Equal[error](t, faultinject.DefaultErr, ctx.Err())
}

func TestInject_withCancelContext(t *testing.T) {
	faultinject.EnableForTest(t)
	expectedErr := random.New(random.CryptoSeed{}).Error()
	type FaultTagFoo struct{}
	ctx := faultinject.Inject(context.Background(), FaultTagFoo{}, expectedErr)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	assert.Equal(t, any(expectedErr), ctx.Value(FaultTagFoo{})) // trigger fault injection
	if ctx.Err() == nil {
		<-ctx.Done()
	}
	assert.ErrorIs(t, expectedErr, ctx.Err())
}

func TestInject_fiFault_ctxErr(t *testing.T) {
	s := testcase.NewSpec(t)
	enabled.Bind(s)

	var (
		parent = testcase.Let[context.Context](s, func(t *testcase.T) context.Context {
			return context.Background()
		})
		fault = testcase.Let[faultinject.CallerFault](s, nil)
	)
	ctx := testcase.Let(s, func(t *testcase.T) context.Context {
		return faultinject.Inject(parent.Get(t), fault.Get(t), exampleErr.Get(t))
	})
	onErr := func(t *testcase.T) error {
		return ctx.Get(t).Err()
	}
	idDoneBlocks := func(t *testcase.T) bool {
		select {
		case <-ctx.Get(t).Done():
			return false
		default:
			return true
		}
	}

	s.When("parent context is cancelled", func(s *testcase.Spec) {
		fault.Let(s, func(t *testcase.T) faultinject.CallerFault {
			return faultinject.CallerFault{Package: "othpkg"}
		})

		parent.Let(s, func(t *testcase.T) context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		})

		s.Then("on .Err, parent. Err is returned", func(t *testcase.T) {
			t.Must.ErrorIs(parent.Get(t).Err(), onErr(t))
		})
	})

	s.When("parent context has an error", func(s *testcase.Spec) {
		fault.Let(s, func(t *testcase.T) faultinject.CallerFault {
			return faultinject.CallerFault{Package: "othpkg"}
		})

		parent.Let(s, func(t *testcase.T) context.Context {
			return StubErrContext{
				Context: context.Background(),
				Error:   t.Random.Error(),
			}
		})

		s.Then("on .Err, parent. Err is returned", func(t *testcase.T) {
			t.Must.ErrorIs(parent.Get(t).Err(), onErr(t))
		})
	})

	s.When("the fi.Fault targets the caller context of the .Err", func(s *testcase.Spec) {
		fault.Let(s, func(t *testcase.T) faultinject.CallerFault {
			return faultinject.CallerFault{
				Package:  "",
				Receiver: "",
				Function: "",
			}
		})

		s.Then("on .Err, the error is returned", func(t *testcase.T) {
			t.Must.ErrorIs(exampleErr.Get(t), onErr(t))
		})

		s.Then(".Done won't block anymore", func(t *testcase.T) {
			t.Must.False(idDoneBlocks(t))
		})

		s.And("after .Err already returned a non-nil error", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				t.Must.NotNil(onErr(t))
			})

			s.Then("successive calls to .Err() return the same error.", func(t *testcase.T) {
				for i, n := 0, t.Random.IntB(3, 7); i < n; i++ {
					t.Must.ErrorIs(exampleErr.Get(t), onErr(t))
				}
			})

			s.Then(".Done won't block anymore", func(t *testcase.T) {
				t.Must.False(idDoneBlocks(t))
			})

			s.Then("checking a value key unrelated to the fault will yield no results", func(t *testcase.T) {
				t.Must.Nil(ctx.Get(t).Value(t.Random.Int()))
			})
		})

		s.And("fault injection is disabled", func(s *testcase.Spec) {
			enabled.LetValue(s, false)

			s.Then("no error is returned", func(t *testcase.T) {
				t.Must.Nil(onErr(t))
			})

			s.Then(".Done will block", func(t *testcase.T) {
				t.Must.True(idDoneBlocks(t))
			})
		})
	})

	s.When("the fi.Fault targeting doesn't match the caller context of the .Err", func(s *testcase.Spec) {
		fault.Let(s, func(t *testcase.T) faultinject.CallerFault {
			return faultinject.CallerFault{
				Package:  "othpkg",
				Receiver: "othReceiver",
				Function: "othFunction",
			}
		})

		s.Then("on .Err, no error is returned", func(t *testcase.T) {
			t.Must.Nil(onErr(t))
		})

		s.Then(".Done will block", func(t *testcase.T) {
			t.Must.True(idDoneBlocks(t))
		})
	})
}

func TestInject_structWithIDField_ctxValue(t *testing.T) {
	s := testcase.NewSpec(t)
	enabled.Bind(s)

	type FooFault struct{ ID any }

	var (
		parent = testcase.Let[context.Context](s, func(t *testcase.T) context.Context {
			return context.Background()
		})
		fault = testcase.Let[FooFault](s, nil)
	)
	ctx := testcase.Let(s, func(t *testcase.T) context.Context {
		return faultinject.Inject(parent.Get(t), fault.Get(t), exampleErr.Get(t))
	})

	var key = testcase.Let[any](s, nil)
	onValue := func(t *testcase.T) any {
		return ctx.Get(t).Value(key.Get(t))
	}

	id := testcase.Let(s, func(t *testcase.T) any {
		return t.Random.Int()
	})
	fault.Let(s, func(t *testcase.T) FooFault {
		return FooFault{ID: id.Get(t)}
	})

	andValueKeyIsSomethingElse(s, onValue, parent, key)

	s.When("the .Value key is the same as the injected Fault", func(s *testcase.Spec) {
		key.Let(s, func(t *testcase.T) any { return FooFault{ID: id.Get(t)} })

		s.Then("an error is returned", func(t *testcase.T) {
			v := onValue(t)
			t.Must.NotNil(v)
			err, ok := v.(error)
			t.Must.True(ok)
			t.Must.NotNil(err)
		})

		andFaultInjectionIsDisabled(s, onValue, parent, key, enabled)
	})

	s.When("the .Value key is the same Fault type but it is not equal by value", func(s *testcase.Spec) {
		key.Let(s, func(t *testcase.T) any { return FooFault{ID: t.Random.Int()} })

		s.Then("on .Value, nil is returned", func(t *testcase.T) {
			t.Must.Nil(onValue(t))
		})

		andFaultInjectionIsDisabled(s, onValue, parent, key, enabled)
	})
}

func andValueKeyIsSomethingElse(s *testcase.Spec,
	onValue func(t *testcase.T) any,
	parent testcase.Var[context.Context],
	key testcase.Var[any],
) {
	s.And(".Value key is something else", func(s *testcase.Spec) {
		key.Let(s, func(t *testcase.T) any {
			return t.Random.String()
		})

		s.And("parent context doesn't have value for the key", func(s *testcase.Spec) {
			s.Then("on .Value, nil is returned", func(t *testcase.T) {
				t.Must.Nil(onValue(t))
			})
		})

		s.And("parent context has a value for the given Key", func(s *testcase.Spec) {
			value := testcase.Let(s, func(t *testcase.T) string {
				return t.Random.String()
			})
			parent.Let(s, func(t *testcase.T) context.Context {
				return context.WithValue(context.Background(), key.Get(t), value.Get(t))
			})
			s.Then("on .Value, the expected value is returned", func(t *testcase.T) {
				t.Must.Equal(value.Get(t), onValue(t))
			})
		})
	})
}

func andFaultInjectionIsDisabled(s *testcase.Spec,
	onValue func(t *testcase.T) any,
	parent testcase.Var[context.Context],
	key testcase.Var[any],
	enabled testcase.Var[bool],
) {
	s.And("fault injection is disabled", func(s *testcase.Spec) {
		enabled.LetValue(s, false)

		s.And("parent context doesn't have value for the key", func(s *testcase.Spec) {
			parent.Let(s, func(t *testcase.T) context.Context {
				return context.Background()
			})

			s.Then("on .Value, nil is returned", func(t *testcase.T) {
				t.Must.Nil(onValue(t))
			})
		})

		s.And("parent context has a value for the given Key", func(s *testcase.Spec) {
			value := testcase.Let(s, func(t *testcase.T) string {
				return t.Random.String()
			})
			parent.Let(s, func(t *testcase.T) context.Context {
				return context.WithValue(context.Background(), key.Get(t), value.Get(t))
			})

			s.Then("on .Value, the expected value is returned", func(t *testcase.T) {
				t.Must.Equal(value.Get(t), onValue(t))
			})
		})
	})
}

type StubErrContext struct {
	context.Context
	Error error
}

func (c StubErrContext) Done() <-chan struct{} {
	if c.Error != nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return c.Context.Done()
}

func (c StubErrContext) Err() error {
	if c.Error != nil {
		return c.Error
	}
	return c.Context.Err()
}
