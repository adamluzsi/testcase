Fault injection allows you to test error test scenarios without adding indirection to your application.
This technique allows you to use concrete types in the same architecture layer and depend less on header-interface-based indirection.

Your client test doesn't have to know about the concrete error values,
but instead can request a fault by a fault injection tag.

Using fault injection across services is also possible. Thus, for example, a client to your API can write integration tests against your system when it encounters an error.
Doing that removes the need to manually maintain mocks at the client-side and leaves more freedom for refactoring.

While fault injection code doesn't cost performance,
you need to judge if header interface-based indirections or fault injection makes more sense for your use-cases.

The Fault injection package doesn't depend on the testing package, and safe to be use in production code. 

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