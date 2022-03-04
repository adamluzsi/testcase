<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [`testcase#Spec`](#testcasespec)
  - [Black-box testing](#black-box-testing)
  - [Variables](#variables)
    - [Usage within a nested scope](#usage-within-a-nested-scope)
  - [Hooks](#hooks)
  - [Basic example with Describe+When+Then](#basic-example-with-describewhenthen)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# `testcase#Spec` 

spec structure is a simple wrapping around the testing.T#Run.
It does not use any global singleton cache object or anything like that.
It does not force you to use global variables.

It uses the same idiom as the core go testing pkg also provide you.
You can use the same way as the core testing pkg
> go run ./... -v -run "TestSomething/the_name_of_the_test_it_print_out_in_case_of_failure"

It allows you to do context preparation for each test in a way,
that it will be safe for use with testing.T#Parallel.

## Black-box testing

I usually only test exported functions, so to me black-box testing worked out the best with specs.
Trough this method, I tend to force myself to create subjects and constructors,
that can be used as examples for the developers who are use my pkg.
> to do black-box testing, just append _test to your current pkg name where you do the testing.

## Variables

in your spec, you can use the `testcase#V` object,
for fetching values for your objects.
Using them is gives you the ability to create value for them,
only when you are in the right testing scope that responsible
for providing an example for the expected value.

To set values, you have to use the [testcase#Spec.Let](https://godoc.org/github.com/adamluzsi/testcase#Spec.Let).
Let will allow you to set a variable to a given scope, and below.
Calling Let in a sub scope will apply the new value for that value to that scope and below.

In test case scopes (`Then`) you will receive a structure ptr called `testcase#V`
which will represent values that you configured for your test case with `Let`.

Values in `testcase#V` are safe to use during T#Parallel execution.

```go
s := testcase.NewSpec(t)

testcase.Let(s, func(t *testcase.T) interface{} {
    return "value"
})

s.Then(`test case`, func(t *testcase.T) {
    t.Log(t.I(`variable name`).(string)) // -> "value"
})
```

### Usage within a nested scope

variables strictly belong to a given `Describe`/`When`/`And` scope,
and configured before any hook would be applied,
therefore hooks always receive the most latest version of the `Let` variable,
regardless where they are defined.

> [Godoc example](https://godoc.org/github.com/adamluzsi/testcase#example-Spec-Let-UsageWithinANestedConext)

if your variable can fail, you can use the T object to assert results before returning the value.

```go
testcase.Let(s, func(t *testcase.T) interface{} {
	t.Fatal(`We can fail let blocks as well, to make sure the let only return consistent values`)

    return "value"
})
```

## [Hooks](/docs/spec/hooks.md)
Hooks allow to setup a certain context with an action,
that needs to be executed before running the testing edge case.
To read more about it, click the link of the documentation, in the section title.

## Basic example with Describe+When+Then

documentation:
* [Context](https://godoc.org/github.com/adamluzsi/testcase#Spec.Context)
* [Test](https://godoc.org/github.com/adamluzsi/testcase#Spec.Test)
* [Describe](https://godoc.org/github.com/adamluzsi/testcase#Spec.Describe)
* [When](https://godoc.org/github.com/adamluzsi/testcase#Spec.When)
* [Then](https://godoc.org/github.com/adamluzsi/testcase#Spec.Then)

```go
package pkgnm

import (
	"testing"
	
	"github.com/adamluzsi/testcase"
	mypkg "path/to/mypkg"
)

func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)

    // when no side effect expected,
    // you can use Spec#Parallel for make all test edge case run on different goroutine
    s.Parallel()

    myType := func(t *testcase.T) *mypkg.MyType {
        return &mypkg.MyType{Field1: t.I(`input`).(string)}
    }

    s.Describe(`IsLower`, func(s *testcase.Spec) {
        subject := func(t *testcase.T) bool { return myType(t).IsLower() }

        s.When(`input string has lower case characters`, func(s *testcase.Spec) {
            testcase.Let(s, func(t *testcase.T) interface{} { return `all lower case` })

            s.Then(`it will return true`, func(t *testcase.T) {
                if subject(t) != true {
                    t.Fatalf(`it was expected that the %q will re reported to be lowercase`, t.I(`input`))
                }
            })

            s.And(`the first character is capitalized`, func(s *testcase.Spec) {
                testcase.Let(s, func(t *testcase.T) interface{} { return `First character is uppercase` })

                s.Then(`it will report false`, func(t *testcase.T) {
                    if subject(t) != false {
                        t.Fatalf(`it was expected that %q will be reported to be not lowercase`, t.I(`input`))
                    }
                })
            })
        })
    })
}
```

