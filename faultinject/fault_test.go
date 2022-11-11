package faultinject_test

import (
	"context"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/faultinject"
	"github.com/adamluzsi/testcase/let"
)

func ExampleCheck() {
	type FaultName struct{}
	ctx := context.Background()

	if err := faultinject.Check(ctx, FaultName{}); err != nil {
		return // err
	}
}

func TestCheck(t *testing.T) {
	s := testcase.NewSpec(t)

	type Key struct{}

	var (
		ctx   = testcase.Let[context.Context](s, nil)
		fault = testcase.LetValue[any](s, Key{})
	)
	act := func(t *testcase.T) error {
		return faultinject.Check(ctx.Get(t), fault.Get(t))
	}

	s.When("context is nil", func(s *testcase.Spec) {
		ctx.LetValue(s, nil)

		s.Then("no error is returned", func(t *testcase.T) {
			t.Must.NoError(act(t))
		})
	})

	s.When("no fault injected", func(s *testcase.Spec) {
		ctx.Let(s, func(t *testcase.T) context.Context {
			return context.Background()
		})

		s.Then("no error is returned", func(t *testcase.T) {
			t.Must.NoError(act(t))
		})
	})

	s.When("fault injected as error", func(s *testcase.Spec) {
		expectedErr := let.Error(s)

		ctx.Let(s, func(t *testcase.T) context.Context {
			return context.WithValue(context.Background(), fault.Get(t), expectedErr.Get(t))
		})

		s.Then("error is returned", func(t *testcase.T) {
			t.Must.ErrorIs(expectedErr.Get(t), act(t))
		})
	})

	s.When("caller fault injected", func(s *testcase.Spec) {
		expectedErr := let.Error(s)
		ctx.Let(s, func(t *testcase.T) context.Context {
			faultinject.EnableForTest(t)
			return faultinject.Inject(
				context.Background(),
				faultinject.CallerFault{},
				expectedErr.Get(t),
			)
		})

		s.Then("error is returned", func(t *testcase.T) {
			t.Must.ErrorIs(expectedErr.Get(t), act(t))
		})

		s.And("the outer context swallow the cancellation", func(s *testcase.Spec) {
			ctx.Let(s, func(t *testcase.T) context.Context {
				d := faultinject.WaitForContextDoneTimeout
				t.Defer(func() { faultinject.WaitForContextDoneTimeout = d })
				faultinject.WaitForContextDoneTimeout = time.Microsecond

				return ctxThatWillNeverGetsDone{Context: ctx.Super(t)}
			})

			s.Then("error is returned after a timeout", func(t *testcase.T) {
				t.Must.ErrorIs(expectedErr.Get(t), act(t))
			})
		})
	})
}

type ctxThatWillNeverGetsDone struct {
	context.Context
}

func (d ctxThatWillNeverGetsDone) Done() <-chan struct{} {
	return make(chan struct{})
}

func (d ctxThatWillNeverGetsDone) Err() error {
	return nil
}

func TestCheck_faultInjectWhenCancelContextTriesToSwallowTheFault(tt *testing.T) {
	t := testcase.NewT(tt, testcase.NewSpec(tt))
	faultinject.EnableForTest(t)
	expectedErr := t.Random.Error()
	callerFault := faultinject.CallerFault{Function: "helperTestCheckFaultInjectWhenCancelContextTriesToSwallowTheFault"}
	ctx := faultinject.Inject(context.Background(), callerFault, expectedErr)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	assert.ErrorIs(t, expectedErr, helperTestCheckFaultInjectWhenCancelContextTriesToSwallowTheFault(ctx))
}

func helperTestCheckFaultInjectWhenCancelContextTriesToSwallowTheFault(ctx context.Context) error {
	type FaultTagFoo struct{}
	return faultinject.Check(ctx, FaultTagFoo{})
}

func ExampleFinish() {
	type FaultName struct{}
	ctx := context.Background()

	_ = func(ctx context.Context) (rErr error) {
		defer faultinject.After(&rErr, ctx, FaultName{})

		return nil
	}(ctx)
}

func TestFinish(t *testing.T) {
	s := testcase.NewSpec(t)

	type (
		Key1 struct{}
		Key2 struct{}
		Key3 struct{}
	)

	var (
		ctx    = testcase.Let[context.Context](s, nil)
		faults = testcase.LetValue[[]any](s, nil)
		rErr   = testcase.Let[*error](s, func(t *testcase.T) *error {
			var err error
			return &err
		})
	)
	act := func(t *testcase.T) {
		faultinject.After(rErr.Get(t), ctx.Get(t), faults.Get(t)...)
	}

	andReturnErrIsNotNil := func(s *testcase.Spec) {
		s.And("return error is not nil", func(s *testcase.Spec) {
			expectedErr := testcase.Let(s, func(t *testcase.T) error {
				return t.Random.Error()
			})
			rErr.Let(s, func(t *testcase.T) *error {
				err := expectedErr.Get(t)
				return &err
			})

			s.Then("error is not modified", func(t *testcase.T) {
				act(t)
				t.Must.ErrorIs(expectedErr.Get(t), *rErr.Get(t))
			})
		})
	}

	s.When("context is nil", func(s *testcase.Spec) {
		ctx.LetValue(s, nil)

		s.Then("no error is returned", func(t *testcase.T) {
			act(t)
			t.Must.Nil(*rErr.Get(t))
		})
	})

	s.When("no fault injected", func(s *testcase.Spec) {
		ctx.Let(s, func(t *testcase.T) context.Context {
			return context.Background()
		})

		s.Then("no error is returned", func(t *testcase.T) {
			act(t)
			t.Must.Nil(*rErr.Get(t))
		})
	})

	s.When("fault injected as error", func(s *testcase.Spec) {
		expectedErr := let.Error(s)
		faults.Let(s, func(t *testcase.T) []any {
			return []any{Key1{}, Key2{}}
		})
		fault := testcase.Let[any](s, nil)
		ctx.Let(s, func(t *testcase.T) context.Context {
			return context.WithValue(context.Background(), fault.Get(t), expectedErr.Get(t))
		})

		s.And("we check for that fault", func(s *testcase.Spec) {
			fault.LetValue(s, Key2{})

			s.Then("error is returned", func(t *testcase.T) {
				act(t)
				t.Must.ErrorIs(expectedErr.Get(t), *rErr.Get(t))
			})

			andReturnErrIsNotNil(s)
		})

		s.And("we don't check for that fault", func(s *testcase.Spec) {
			fault.LetValue(s, Key3{})

			s.Then("no error is returned", func(t *testcase.T) {
				act(t)
				t.Must.NoError(*rErr.Get(t))
			})
		})
	})

	s.When("caller fault injected", func(s *testcase.Spec) {
		expectedErr := let.Error(s)
		faults.LetValue(s, nil)
		ctx.Let(s, func(t *testcase.T) context.Context {
			faultinject.EnableForTest(t)
			return faultinject.Inject(
				context.Background(),
				faultinject.CallerFault{},
				expectedErr.Get(t),
			)
		})

		s.Then("error is returned", func(t *testcase.T) {
			act(t)

			t.Must.ErrorIs(expectedErr.Get(t), *rErr.Get(t))
		})

		andReturnErrIsNotNil(s)
	})
}
