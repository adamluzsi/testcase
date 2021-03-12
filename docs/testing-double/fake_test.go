package testingdouble_test

import (
	"context"
	"testing"

	"github.com/adamluzsi/testcase/fixtures"
)

type FakeXYStorage map[string]XY

func (f FakeXYStorage) CreateXY(ctx context.Context, ptr *XY) error {
	if ptr.ID == `` {
		ptr.ID = fixtures.Random.StringN(42) // not safe
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
		Fixtures: FixtureFactory{},
	}.Test(t)
}
