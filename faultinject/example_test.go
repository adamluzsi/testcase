package faultinject_test

import (
	"context"
	"errors"
	"fmt"

	"go.llib.dev/testcase/faultinject"
	"go.llib.dev/testcase/random"
)

type (
	FaultTag struct{}
)

func Example() {
	defer faultinject.Enable()()
	ctx := context.Background()
	fmt.Println(ctx.Err()) // no error as expected

	// arrange one fault injection for FaultTag
	ctx = faultinject.Inject(ctx, FaultTag{}, fmt.Errorf("example error to inject"))

	if err, ok := ctx.Value(FaultTag{}).(error); ok {
		fmt.Println(err) // prints the injected error
	}
	if err, ok := ctx.Value(FaultTag{}).(error); ok {
		fmt.Println(err) // code not reached as injectedFault is already consumed
	}
}

func ExampleAfter() {
	type fault struct{}
	ctx := faultinject.Inject(context.Background(), fault{}, fmt.Errorf("boom"))

	_ = func(ctx context.Context) (returnErr error) {
		defer faultinject.After(&returnErr, ctx, fault{})

		return nil
	}(ctx)
}

func Example_chaosEngineeringWithExplicitFaultPoints() {
	defer faultinject.Enable()()
	ctx := context.Background()
	fmt.Println(MyFuncWithChaosEngineeringFaultPoints(ctx)) // no error

	ctx = faultinject.Inject(ctx, FaultTag{}, errors.New("boom")) // arrange fault injection for FaultTag
	fmt.Println(MyFuncWithChaosEngineeringFaultPoints(ctx))       // "boom" is returned
}

func MyFuncWithChaosEngineeringFaultPoints(ctx context.Context) error {
	if err, ok := ctx.Value(FaultTag{}).(error); ok {
		return err
	}

	// check for injected Fault that target this caller stack
	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}

func Example_faultInjectWithFixErrorReplyFromTheFaultPoint() {
	defer faultinject.Enable()()
	ctx := context.Background()
	ctx = faultinject.Inject(ctx, FaultTag{}, errors.New("ignored"))
	fmt.Println(MyFuncWithFixErrorReplyFromTheFaultPoin(ctx)) // error is returned
}

func MyFuncWithFixErrorReplyFromTheFaultPoin(ctx context.Context) error {
	if _, ok := ctx.Value(FaultTag{}).(error); ok {
		return errors.New("my error value")
	}

	return nil
}

func ExampleInject_byTargetingCaller() {
	ctx := faultinject.Inject(
		context.Background(),
		faultinject.CallerFault{
			Package:  "", // empty will match everything
			Receiver: "", //
			Function: "", //
		},
		errors.New("boom"),
	)

	fmt.Println(ctx.Err()) // "boom"
}

func ExampleCallerFault() {
	defer faultinject.Enable()()
	ctx := context.Background()
	fmt.Println(MyFuncWithStandardContextErrCheck(ctx)) // no error

	fault := faultinject.CallerFault{ // inject Fault that targets a specific context Err check
		Function: "MyFuncWithStandardContextErrCheck", // selector to tell where to inject
	}
	err := random.New(random.CryptoSeed{}).Error()

	ctx = faultinject.Inject(ctx, fault, err)           // some random error)
	fmt.Println(MyFuncWithStandardContextErrCheck(ctx)) // Fault.Error injected and returned
}

func MyFuncWithStandardContextErrCheck(ctx context.Context) error {
	if err := ctx.Err(); err != nil { // go idiom to cancel processing on bad context
		return err
	}

	return nil
}
