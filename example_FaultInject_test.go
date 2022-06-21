package testcase_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/adamluzsi/testcase/faultinject"
)

type (
	FaultTag1 struct{}
	FaultTag2 struct{}
	FaultTag3 struct{}
)

func Example_faultInject() {
	defer faultinject.Enable()()
	ctx := context.Background()
	// arrange fault injection for my-tag-1
	ctx = faultinject.Inject(ctx, FaultTag1{})
	// no error
	fmt.Println(fii.Check(context.Background()))
	// yields error
	fmt.Println(fii.Check(ctx))
}

var fii = faultinject.Injector{}.
	OnTag(FaultTag1{}, errors.New("boom1")).
	OnTag(FaultTag2{}, errors.New("boom2")).
	OnTag(FaultTag3{}, errors.New("boom3"))

func MyFunc(ctx context.Context) error {
	// single tag checking case
	if err := fii.CheckFor(ctx, FaultTag2{}); err != nil {
		return err
	}

	// check for any registered tag
	if err := fii.Check(ctx); err != nil {
		return err
	}

	return nil
}
