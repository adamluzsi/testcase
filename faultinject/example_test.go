package faultinject_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/adamluzsi/testcase/faultinject"
)

type (
	Tag1 struct{}
	Tag2 struct{}
	Tag3 struct{}
)

func Example() {
	ctx := context.Background()
	// arrange fault injection for my-tag-1
	ctx = faultinject.Inject(ctx, Tag1{})
	// no error
	fmt.Println(fii.Check(context.Background()))
	// yields error
	fmt.Println(fii.Check(ctx))
}

var fii = faultinject.Injector{}.
	OnTag(Tag1{}, errors.New("boom1")).
	OnTag(Tag2{}, errors.New("boom2")).
	OnTag(Tag3{}, errors.New("boom3"))

func MyFunc(ctx context.Context) error {
	if err := fii.Check(ctx); err != nil {
		return err
	}

	return nil
}
