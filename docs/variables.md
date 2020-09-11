# Variables in specifications

## Summary
In `testcase` you can use two approach to manage variables in your specs:
* The suggested approach is to use specify variables with `Spec#Let`/`Spec#LetValue`/`T#Let`.
* As an alternative way you can use variable scope injection and `Spec#Before` hook to manipulate the content of the variable.

## Context 
During spec definition, the spec description often starts from high level, 
and then genuinely specify the context of the test case with each sub testing scope.

When we have a test subject, we often provide some form of variables to it,
that we later define with a value in a subcontext.
This helps in building specs, where we can fine tune the inputs for the test subject. 

In most framework, the only and most traditional way to manage variables,
is through injecting the scope of a variable into subcontexts.
We define a variable before the test subject,
then we alter the content in a subcontext with the help of testing hooks.
  
This usually forces variables to be reused between tests,
and without further heavy workarounds,
the individual tests in the spec should not run concurrently,
even if the test subject itself is not expected to have any side effects.

This also has the pitfall that you have to remember which variable is properly configured in a testing scope,
since go initialize types with zero values.

If you have multiple variables for a test subject, you have to keep a bigger mental model about the specification,
and manually ensure, that each variable in the subcontext is properly set with the right values.

You also need to ensure that if you set a variable to a certain value in a scope,
this doesn't leak out and being implicitly used by another test
which is actually located outside of the current testing scope.

```go
package xyz_test

import (
    "testing"
    . "someframework"
)

func TestMyFunction(t *testing.T) {
    var (
        a string
        b int
    )
    subject := func() error {
        return MyFunction(a, b)
    }
    Context(`when a...`, func() {
        Before(func(t *testing.T){
            a = "foo"
        })

        Context(`and b...`, func() {
            Before(func(t *testing.T) {
                b = 42
            })

            Test(t, `then should accept`, func(t *testing.T) {
                if err := subject(); err != nil {
                    t.FailNow()
                }
            })      
        })
    })
    Context(`when a...`, func() {
        Before(func(t *testing.T){
            a = "bar"
        })
   
        // accidentally leaked b value here
        Test(t, `then should raise an error`, func(t *testing.T) {
            if err := subject(); err == nil {
                t.FailNow()
            }
        })   
  })
}  
```
> pseudo example

## Solution
In `testcase`, you provided with a set of tool to manage variables which bound to the currently running test case lifetime.
You can reference variables on top contexts and define them later 
when you actually focus on describing different behaviors with different inputs regarding a variable.

If you forgot to set a variable, the spec will warn you about the undefined variable usage,
and you can think trough if you forgot to specify something in your current testing scope.

To define variables, you can use `Spec#Let`/`Spec#LetValue`/`T#Let`.
To access variables, you can use `T#I` together with interface casting.

Let define a memoized helper method.
Let creates lazily-evaluated test execution bound variables.
Let variables don't exist until called into existence by the actual tests,
so you won't waste time loading them for examples that don't use them.
This allows you to have domain interactros with they resource dependency defined in `Let` variables in a helper function,
and only initialize and cleanup in tests where you actually use them as a dependency.

This allows to build project specific testing suites where dependency inejection in tests can be easily maintained.
Imagine you make a http handler test, and you can simply use real domain use-case instances to prepare your test context.
And if you need to change how a domain use-case is initialized or torn down, you can do that in one place.

`Let` variables also memoized, so they're useful for encapsulating database objects, due to the cost of making a database request.
The value will be cached across all use within the same test execution but not across different test cases.
You can eager load a value defined in let by referencing to it in a Before hook.

Since variables defined through this will belong to the test case runtime,
when no side effect can be observed with the test subject,
it becomes safe to test them concurrently.
In other words, `Let` is threadsafe, and each test, including parallel ones will receive they own test variable instance.

For examples on how to use this, please check the godoc examples.
- [basic usage of Let](https://pkg.go.dev/github.com/adamluzsi/testcase?tab=doc#example-Spec.Let)
- [define immutable values without a block using LetValue](https://pkg.go.dev/github.com/adamluzsi/testcase?tab=doc#example-Spec.LetValue)
- [managing mock lifecycle with a let variable](https://pkg.go.dev/github.com/adamluzsi/testcase?tab=doc#example-Spec.Let-Mock)
- [usage of SqlDB with a let variable](https://pkg.go.dev/github.com/adamluzsi/testcase?tab=doc#example-Spec.Let-SqlDB)
- [Usage Within Nested Scope](https://pkg.go.dev/github.com/adamluzsi/testcase?tab=doc#example-Spec.Let-UsageWithinNestedScope)

### Why use Let?
We’ve established that standard variable declaration with scope injection through `Before` hooks,
and `Let` have different characteristics.
So what are the advantages of using `Let`?
`Let` helps to DRY up your tests. 

You can have the same value actross all of your examples without actually sharing the same instance of the value.
You can also override thing to a different value, in a subcontext block.

Lazily evaluated. `Before` hooks can make your tests slow when they set up a lot of state
because the whole before block gets called even when running a single test.
If you use `Let`, instead, then it would only set up the state required for the specific test that you’re running.

`Let` can make your tests easier to read for simple examples.
If you look at the examples above, you’ll might agree that the test implementation that uses `Let` is simpler,
contains less code and is easier to read than the before block alternative.

### The problem with Let
Using `Let` usually turns out well with small spec files. 
Problems start when your spec files become large.
When you use a let in the top-level describe block it becomes global to all of your tests.

If you have a large spec file it can be difficult to know what is in scope 
because a `Let` can be far away from the code that you’re looking at.

It becomes difficult to tell an origin of a commonly shared variables from spec helpers
without some form of accessor function, 
that gives a hint to the reader where the variable prepared for the test.

## Consequence
PRO:
- variables less likely to leak out due to mistakes during the write of the specification
- thread safe way to "share" variables between test examples. 
- increased local development feedback loop because side effect free tests can be executed concurrently.
- specs can be written in a more dry way,
    where details can be added later without overwriting the preparation hooks on the given testing context. 
- forgotten variable decelerations will cause a warning to the software engineer,
    and request them to make a conscious decision on what value should be used in the given scope.
- lazy loaded values allow reusability with spec helper setup functions.
    *  dependency injection becomes much more maintainable through the use of spec helpers where common variables defined.

CON:
- the learning curve of the test increases
- requires a slightly different mindset during testing specification.

## External references
- [rspec](https://github.com/rspec/rspec)
- [better specs about let](https://www.betterspecs.org/#let)
- [medium post about RSpec let by Tom Kadwill](https://medium.com/@tomkadwill/all-about-rspec-let-a3b642e08d39)
