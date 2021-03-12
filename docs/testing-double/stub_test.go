package testingdouble_test

import "context"

type StubXYStorage struct {
	CreateXYFunc   func(ctx context.Context, ptr *XY) error
	FindXYByIDFunc func(ctx context.Context, ptr *XY, id string) (found bool, err error)
}

func (stub StubXYStorage) CreateXY(ctx context.Context, ptr *XY) error {
	return stub.CreateXYFunc(ctx, ptr)
}

func (stub StubXYStorage) FindXYByID(ctx context.Context, ptr *XY, id string) (found bool, err error) {
	return stub.FindXYByIDFunc(ctx, ptr, id)
}

var _ XYStorage = StubXYStorage{}
