<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [testcase nesting guide](#testcase-nesting-guide)
  - [Flattening](#flattening)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# testcase nesting guide

In `testcase` to express certain edge cases,
the framework prefers the usage of nesting.

By convention every `if` statement should have 2 corresponding testing context to represent possible edge cases.
This is required in order to keep clean track of the code complexity.
If the test coverage became too big or have too many level of nesting, 
that is the clear sign that the implementation has too broad scope,
and the required mental model for the given production code code is likely to be big.

* [example code](/docs/examples/ValidateName.go)
* [example test](/docs/examples/ValidateName_test.go)

For implementations where you need to test business logic, 
`testcase#Spec` is suggested, even if the spec has too many nested layers.
That is only represent the complexity of the component.

## Flattening

When the specification becomes too big,
you can improve readability by flattening the specification
by refactoring out specification sub-context(s) into a function.

The added benefit for this is that all the common variables present on the top level
can be accessed from each of the sub context.
For e.g you can create variable with a database connection,
that ensure a common way to connect and afterwards cleaning up the connection.  

```go
package examples_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/docs/examples"
)

func TestMyStruct(t *testing.T) {
	s := testcase.NewSpec(t)
	s.NoSideEffect()

	testcase.Let(s, func(t *testcase.T) interface{} {
		return examples.MyStruct{}
	})

	// define shared variables and hooks here
	// ...

	s.Describe(`Say`, SpecMyStruct_Say)
	s.Describe(`Foo`, SpecMyStruct_Foo)
	// other specification sub contexts
}

func SpecMyStruct_Say(s *testcase.Spec) {
	var subject = func(t *testcase.T) string {
		return t.I(`my-struct`).(examples.MyStruct).Say()
	}

	s.Then(`it will say a famous quote`, func(t *testcase.T) {
		assert.Must(t).Equal( `Hello, World!`, subject(t))
	})
}

func SpecMyStruct_Foo(s *testcase.Spec) {
	var subject = func(t *testcase.T) string {
		return t.I(`my-struct`).(examples.MyStruct).Foo()
	}

	s.Then(`it will say a famous quote`, func(t *testcase.T) {
		assert.Must(t).Equal( `Bar`, subject(t))
	})
}
```

* [example code](/docs/examples/MyStruct.go)
* [example test](/docs/examples/MyStruct_test.go)
