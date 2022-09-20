package faultinject_test

import (
	"context"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/faultinject"
)

var enabled = testcase.Var[bool]{
	ID: "faultinject is enabled",
	Init: func(t *testcase.T) bool {
		return true
	},
	OnLet: func(s *testcase.Spec, enabled testcase.Var[bool]) {
		s.Before(func(t *testcase.T) {
			if enabled.Get(t) {
				faultinject.EnableForTest(t)
			}
		})
	},
}

var exampleErr = testcase.Var[error]{
	ID: "example error",
	Init: func(t *testcase.T) error {
		return t.Random.Error()
	},
}

type ExampleReceiver struct {
	MainRanFaultPoint bool
	MainIsFinished    bool
}

func (r *ExampleReceiver) Main(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		r.MainRanFaultPoint = true
		return err
	}
	if err := r.OnErr(ctx); err != nil {
		return err
	}
	if err := r.OnValue(ctx); err != nil {
		return err
	}
	r.MainIsFinished = true
	return nil
}

func (r *ExampleReceiver) OnErr(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (r *ExampleReceiver) OnValue(ctx context.Context) error {
	type SomeTag struct{}
	if err, ok := ctx.Value(SomeTag{}).(error); ok {
		return err
	}
	return nil
}
