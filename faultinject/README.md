# Fault Injection

Fault injection is a chaos engineering utility
that allows you to test error scenarios
with real components by adding fault points.

The package has strictly tested.

## Features

- You can add fault points to specific points
    - This simplifies your expectations in your integration tests; you can trigger these fault points instead of
      analyzing the underlying error handling of a given component
      when you don't want the test of the implementation details of a given dependent component
- The ability to simulate temporary errors
    - This allows testing retry logic with ease without the need to build and maintain intelligent mocks that try to
      mimic real components
- Global Enable switch to allow fault injection on demand
    - By default, fault injection ignores all calls, doesn't check, doesn't inject unless Fault Injection is explicitly
      allowed

## Example

The Fault injection package doesn't depend on the testing package and should be safe to use in production code.

```go
package mypkg

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

func main() {
	defer faultinject.Enable()()
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
```

## Description

This approach enables you to test out small error cases or event the cascading effects in a microservice setup.
One of the instant benefits of fault injection is that your clients can test with your actual errors
and don't need to maintain their mocks/stubs arrangements manually.
If fuel injection is exposed on your API, then It also enables your clients to write integration tests against error
scenarios with your system's API.
Last but not least, it allows you to remove forced indirections from your codebase,
where you have to use a header interface for the sake of testing error handling in a component.

One often mentioned argument about fault injection is the need to add something to the production codebase for testing,
but in practice, if you have many header interfaces in your codebase, then you are already actively altering your
production codebase for testing purposes,
In the end, you need to judge if header interface-based indirections or fault injection makes more sense for your
use-cases, as this is not a silver bullet.

By this time, I believe you might feel reticent to put fault injection into your non-test code.
Engineering controlled chaos into your application is not a standard testing strategy.
It has its pros and cons. For example, you can simplify your code
through using less header interface based indirection to test error cases in your code.
It allows you to trigger faults without the need to understand
the internal logic of that concrete implementation's error cases.
It also allows you to specify expectations about fault injection in your Role Interface's interface testing suite.
You can do simulation of temporary outages and test retry mechanisms.

But on the grand scale, the real value with fault injection is the ability
to test error cases at the system level in a micro-service setup,
where errors can have unexpected cascading effects.
I loved to see how easy to find bugs with fault injection
when I was working with a mobile team in one of my previous job.
Our biggest issue was that after you released a mobile client, it was a pain point to make our users upgrade to the
latest client version when we identified rainy cases during production use.
