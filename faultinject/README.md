# Fault Injection

Fault injection is a chaos engineering utility that allows you to test error scenarios with real components by adding fault points.
This approach enables you to test out small error cases or event the cascading effects in a microservice setup.
One of the instant benefits of fault injection is that your clients can test with your actual errors
and don't need to maintain their mocks/stubs arrangements manually.
If fuel injection is exposed on your API, then It also enables your clients to write integration tests against error scenarios with your system's API.
Last but not least, it allows you to remove forced indirections from your codebase,
where you have to use a header interface for the sake of testing error handling in a component.

One often mentioned argument about fault injection is the need to add something to the production codebase for testing,
but in practice, if you have many header interfaces in your codebase, then you are already actively altering your production codebase for testing purposes,
In the end, you need to judge if header interface-based indirections or fault injection makes more sense for your use-cases, as this is not a silver bullet.

The Fault injection package doesn't depend on the testing package and is safe to be used in production code.

```go
package main

import (
	"context"
	"errors"

	"github.com/adamluzsi/testcase/faultinject"
)

func main() {
	ctx := context.Background()
	ctx = faultinject.Inject(ctx, "my-tag-2")

	_ = MyFunc(ctx) // boom2
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
```