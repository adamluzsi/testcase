package faultinject_test

import (
	"context"
	"errors"

	"github.com/adamluzsi/testcase/faultinject"
)

func MyFunc(ctx context.Context) error {
	if err := faultinject.Check(ctx, "my-tag"); err != nil {
		return err
	}

	// your logic goes here
	return nil
}

func ExampleCheck() {
	ctx := context.Background()

	_ = MyFunc(ctx) // no error

	ctx = faultinject.Inject(ctx, faultinject.Fault{
		OnFunc: "faultinject_test.MyFunc",
		Error:  errors.New("boom1"),
	})
	ctx = faultinject.Inject(ctx, faultinject.Fault{
		OnTag: "my-tag",
		Error: errors.New("boom2"),
	})

	_ = MyFunc(ctx) // yields error -> boom1
	_ = MyFunc(ctx) // yields error -> boom2
	_ = MyFunc(ctx) // no error
}

func ExampleInject() {
	ctx := context.Background()

	_ = MyFunc(ctx) // no error

	ctx = faultinject.Inject(ctx, faultinject.Fault{
		OnFunc: "faultinject_test.MyFunc",
		Error:  errors.New("boom1"),
	})
	ctx = faultinject.Inject(ctx, faultinject.Fault{
		OnTag: "my-tag",
		Error: errors.New("boom2"),
	})

	_ = MyFunc(ctx) // yields error -> boom1
	_ = MyFunc(ctx) // yields error -> boom2
	_ = MyFunc(ctx) // no error
}
