package mypkg

import (
	"context"
)

type ExampleStruct struct {
	MainRanFaultPoint bool
	MainIsFinished    bool
}

func (r *ExampleStruct) Main(ctx context.Context) error {
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

func (r *ExampleStruct) OnErr(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (r *ExampleStruct) OnValue(ctx context.Context) error {
	type SomeTag struct{}
	if err, ok := ctx.Value(SomeTag{}).(error); ok {
		return err
	}
	return nil
}
