package testingdouble_test

import "context"

type StubMethodCreateXY struct {
	XYStorage

	CreateXYFunc func(ctx context.Context, ptr *XY) error
}

func (stub StubMethodCreateXY) CreateXY(ctx context.Context, ptr *XY) error {
	if stub.CreateXYFunc != nil {
		return stub.CreateXYFunc(ctx, ptr)
	}

	return stub.XYStorage.CreateXY(ctx, ptr)
}

var _ XYStorage = StubMethodCreateXY{}
