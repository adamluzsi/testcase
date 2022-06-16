package faultinject_test

import (
	"context"
	"errors"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/faultinject"
)

func Example() {
	ctx := context.Background()
	// arrange fault injection for my-tag-1
	ctx = faultinject.Inject(ctx, "my-tag-1")

	var tb testing.TB
	assert.ErrorIs(tb, errors.New("boom1"), MyFunc(ctx))
}

var fii = faultinject.Injector{}.
	OnTag("my-tag-1", errors.New("boom1")).
	OnTag("my-tag-2", errors.New("boom2")).
	OnTag("my-tag-2", errors.New("boom3"))

func MyFunc(ctx context.Context) error {
	if err := fii.Check(ctx); err != nil {
		return err
	}

	return nil
}
