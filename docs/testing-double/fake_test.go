package testingdouble_test

import (
	"context"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/random"
)

type FakeXYStorage map[string]XY

func (f FakeXYStorage) CreateXY(ctx context.Context, ptr *XY) error {
	rnd := random.New(random.CryptoSeed{})
	if ptr.ID == `` {
		ptr.ID = rnd.StringN(42) // not safe
	}
	f[ptr.ID] = *ptr
	return nil
}

func (f FakeXYStorage) FindXYByID(ctx context.Context, ptr *XY, id string) (found bool, err error) {
	ent, ok := f[id]
	if !ok {
		return false, nil
	}
	*ptr = ent
	return true, nil
}

// file: FakeXYEntityStorage_test.go

var _ XYStorage = FakeXYStorage{}

func TestFakeXYEntityStorage_suppliesXYStorageContract(t *testing.T) {
	XYStorageContract{
		Subject: func(tb testing.TB) XYStorage {
			return make(FakeXYStorage)
		},
		MakeCtx: func(tb testing.TB) context.Context {
			return context.Background()
		},
		MakeXY: func(tb testing.TB) *XY {
			t := testcase.NewT(tb, nil)
			return t.Random.Make(new(XY)).(*XY)
		},
	}.Test(t)
}
