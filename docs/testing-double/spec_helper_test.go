package testingdouble_test

import (
	"context"
	"testing"

	"github.com/adamluzsi/testcase/fixtures"
)

type XY struct {
	ID string
	V  int
}

// Consumer is the business use-case that depends on a XYStorage role.
type Consumer struct {
	Storage XYStorage
}

func (c Consumer) DoSomething(ctx context.Context) {
	// use XYStorage here
}

// XYStorage is the role interface
type XYStorage interface {
	CreateXY(ctx context.Context, ptr *XY) error
	FindXYByID(ctx context.Context, ptr *XY, id string) (found bool, err error)
}

// ./contracts package

type XYStorageContract struct {
	Subject  func(tb testing.TB) XYStorage
	Fixtures interface {
		CreateXYEntity() *XY
		Context() context.Context
	}
}

func (c XYStorageContract) Test(t *testing.T) {
	// test behaviour expectations about the storage methods
	t.Run(`when entity created in storage, it should assign ID to the received entity and the entity should be located in the storage`, func(t *testing.T) {
		var (
			subject = c.Subject(t)
			ctx     = c.Fixtures.Context()
			entity  = c.Fixtures.CreateXYEntity()
		)

		if err := subject.CreateXY(ctx, entity); err != nil {
			t.Fatal(`XYStorage.Create failed:`, err.Error())
		}

		id := entity.ID

		if id == `` {
			t.Fatal(`XY.ID was expected to be populated after CreateXY is called`)
		}

		t.Log(`entity should be findable in the storage after Create`)

		var actual XY

		found, err := subject.FindXYByID(ctx, &actual, id)
		if err != nil {
			t.Fatal(`XYStorage.FindByID failed:`, err.Error())
		}
		if !found {
			t.Fatal(`it was expected that entity can be found in the storage by id`)
		}

		if actual != *entity {
			t.Fatal(`it was expected that stored entity is the same as the one being persisted in the storage`)
		}
	})
}

func (c XYStorageContract) Benchmark(b *testing.B) {
	// benchmark
	b.SkipNow()
}

type FixtureFactory struct{}

func (ff FixtureFactory) CreateXYEntity() *XY {
	return fixtures.New(XY{}, fixtures.SkipByTag(`ID`)).(*XY)
}

func (ff FixtureFactory) Context() context.Context {
	return context.Background()
}
