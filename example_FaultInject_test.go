package testcase_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/adamluzsi/testcase/faultinject"
)

type (
	FaultTag          struct{}
	FaultTagWithError struct{ Error error }
)

func Example_faultInject() {
	disableFaultInject := faultinject.Enable()
	defer disableFaultInject()
	// or defer faultinject.Enable()()

	err := errors.New("my injected error")
	ctx := context.Background()
	ctx = faultinject.Inject(ctx, FaultTag{}, nil)                                                    // with default err
	ctx = faultinject.Inject(ctx, FaultTag{}, err)                                                    // with injected err
	ctx = faultinject.Inject(ctx, faultinject.CallerFault{Function: "MyFuncWithFaultInjection"}, err) // fault inject by caller

	fmt.Println(MyFuncWithFaultInjection(ctx)) // yields back error because FaultTag{}
	fmt.Println(MyFuncWithFaultInjection(ctx)) // yields back the provided error with FaultTagWithError
	fmt.Println(MyFuncWithFaultInjection(ctx)) // yields back the provided error with faultinject.Fault
	fmt.Println(MyFuncWithFaultInjection(ctx)) // no more error
}

func MyFuncWithFaultInjection(ctx context.Context) error {
	// check for fault injection
	if err := ctx.Err(); err != nil {
		return err
	}

	// or check for a specific error tag with default fault injection provided error
	if ctx.Value(FaultTag{}) != nil {
		return errors.New("an error that we define since we know the best what makes sense as an error from here")
	}

	// or if you want to inject error from your test
	if err, ok := ctx.Value(FaultTag{}).(error); ok {
		return err
	}

	return nil
}

func Example_faultInjectWithContextErr() {
	defer faultinject.Enable()()

	ctx := context.Background()

	// all fault field is optional.
	// in case left as zero value,
	// it will match every caller context,
	// and returns on the first .Err() / .Value(faultinject.Fault{})
	ctx = faultinject.Inject(ctx, faultinject.CallerFault{
		Package:  "targetpkg",
		Receiver: "*myreceiver",
		Function: "myfunction",
	}, errors.New("boom"))

	// from and after call stack reached: targetpkg.(*myreceiver).myfunction
	if err := ctx.Err(); err != nil {
		fmt.Println(err) // in the position defined by the Fault, it will yield an error
	}
}
