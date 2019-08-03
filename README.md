<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [testcase GoDoc](#testcase-godoc)
  - [The reason behind the package](#the-reason-behind-the-package)
  - [What makes testcase different ?](#what-makes-testcase-different-)
  - [Reference Projects](#reference-projects)
  - [The Spec based approach](#the-spec-based-approach)
    - [Black-box testing](#black-box-testing)
    - [Variables](#variables)
      - [Usage within a nested scope](#usage-within-a-nested-scope)
    - [Hooks](#hooks)
      - [Before](#before)
      - [After](#after)
      - [Around](#around)
    - [Basic example with Describe+When+Then](#basic-example-with-describewhenthen)
  - [The Steps struct based approach](#the-steps-struct-based-approach)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![GoDoc](https://godoc.org/github.com/adamluzsi/testcase?status.png)](https://godoc.org/github.com/adamluzsi/testcase)
[![Build Status](https://travis-ci.org/adamluzsi/testcase.svg?branch=master)](https://travis-ci.org/adamluzsi/testcase)
[![Go Report Card](https://goreportcard.com/badge/github.com/adamluzsi/testcase)](https://goreportcard.com/report/github.com/adamluzsi/testcase)
[![codecov](https://codecov.io/gh/adamluzsi/testcase/branch/master/graph/badge.svg)](https://codecov.io/gh/adamluzsi/testcase)
# testcase

The package considered stable and no changes expected to the package exported API.

The main documentation is kept in the [GoDoc](https://godoc.org/github.com/adamluzsi/testcase),
and this README serves only as a high level introduction.

This package implements two approaches to help you to do nested BDD style testing in golang.

The package may seems inactive maybe, but it is used daily,
I just don't plan to feature creep it,
because it is totally efficient to achieve what I need,
and If I need extra helper function or anything like that,
I usually put it under the $PROJECT_ROOT/testing package.
Then I include the helpers with `.` importing.
I highly discourage the use of the  dot notation based import outside of the testing files.

## My totally Biased Opinion about this project

Primary I made this project for myself,
because using vanilla`testing#T.Run` forced me to apply repetitive boilerplate
in every test, and I wanted to introduce some form of maintainability for my tests.
I want to stick as much as possible with the core testing pkg,
so this mainly just to have those boilerplates in the form of centralized package.

I normally okay with my creations,
but I really really love this project,
because it give me a huge productivity boost,
and also it helps to apply my convention for testing.
It may not for everyone, and that is totally fine.
There are tons of testing frameworks out there,
with huge community support.

Also I need to mention, that this project is heavily based on the experience I made working with [rspec](https://github.com/rspec/rspec).
I highly recommend checking out that project and the [community takeaways about how to write a better software specification](http://www.betterspecs.org).

I don't plan on doing complex custom things in this package.
For example I don't plan to have a visually appealing reporting output
or custom assertion helpers.
No, kind the opposite, since the output intentionally looks like vanilla `testing` run output.
I need the ability to keep things close to core go testing pkg conventions,
so I can use things like `-run 'rgx'` flag.

Therefore again this project is here for my own work primary,
but please feel free to use it if you see value in it for yourself.

The project only goal is to make it easy and productive to create isolated test cases,
reproducible setup/teardown logic
and testing context based variable scoping.

## How much this project will be maintained ?

This project is based on the `testing` package [T.Run](https://godoc.org/testing#T.Run) *idiom*,
so basically as long that is supported and maintained by the golang core team,
this project is easily considered up to date.

I use it for my private projects,
but I designed this project to be cost effective for my time.
I only piggybacking the core golang team work basically.

## [The reason behind the package](https://godoc.org/github.com/adamluzsi/testcase#hdr-The_reason_behind_the_package)
## [What makes testcase different ?](https://godoc.org/github.com/adamluzsi/testcase#hdr-What_makes_testcase_different)

## Reference Project
* [toggler](https://github.com/adamluzsi/toggler)

## The Spec based approach

spec structure is a simple wrapping around the testing.T#Run.
It does not use any global singleton cache object or anything like that.
It does not force you to use global variables.

It uses the same idiom as the core go testing pkg also provide you.
You can use the same way as the core testing pkg
> go run ./... -v -run "the/name/of/the/test/it/print/out/in/case/of/failure"

It allows you to do context preparation for each test in a way,
that it will be safe for use with testing.T#Parallel.

### Vanilla `testing#T.Run` like approach

```go
func TestMyType(t *testing.T) {
	s := testcase.NewSpec(t)

	myType := func(t *testcase.T) *MyType {
		return &MyType{Field1: `input`}
	}

	s.Run(`describe IsLower`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) bool { return myType(t).IsLower() }

		s.Context(`when something`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { /* setup */ })

			s.Test(`test-case`, func(t *testcase.T) {
				require.True(t, subject(t))
			})
		})
	})
}
```

### Black-box testing

I usually only test exported functions, so to me black-box testing worked out the best with specs.
Trough this method, I tend to force myself to create subjects and constructors,
that can be used as examples for the developers who are use my pkg.
> to do black-box testing, just append _test to your current pkg name where you do the testing.

### Variables

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

s.Let(`variable name`, func(t *testcase.T) interface{} {
    return "value"
})

s.Then(`test case`, func(t *testcase.T) {
    t.Log(t.I(`variable name`).(string)) // -> "value"
})
```

#### Usage within a nested scope

variables strictly belong to a given `Describe`/`When`/`And` scope,
and configured before any hook would be applied,
therefore hooks always receive the most latest version of the `Let` variable,
regardless where they are defined.

> [Godoc example](https://godoc.org/github.com/adamluzsi/testcase#example-Spec-Let-UsageWithinANestedConext)

if your variable can fail, you can use the T object to assert results before returning the value.

```go
s.Let(`input`, func(t *testcase.T) interface{} {
	t.Fatal(`We can fail let blocks as well, to make sure the let only return consistent values`)

    return "value"
})
```

### Hooks

Hooks help you setup common things for each test case.
For example clean ahead, clean up, mock expectation configuration,
and similar things can be done in hooks,
so your test case blocks with `Then` only represent the expected result(s).

In case you work with something that depends on side-effects,
such as database tests, you can use the hooks,
to create clean-ahead / clean-up blocks.

Also if you use gomock, you can use the spec#Around function,
to set up the mock with a controller, and in the teardown function,
call the gomock.Controller#Finish function,
so your test cases will be only about
what is the different behavior from the rest of the test cases.

It will panic if you use hooks or variable preparation in an ambiguous way,
or when you try to access variable that doesn't exist in the context where you do so.
It tries to panic with friendly and supportive messages, but that is highly subjective.

#### Before

Before give you the ability to run a block before each test case.
This is ideal for doing clean ahead before each test case.
The received *testing.T object is the same as the Then block *testing.T object
This hook applied to this scope and anything that is nested from here.
All setup block is stackable.

```go
s := testcase.NewSpec(t)

s.Before(func(t *testcase.T) {
    // this will run before the test cases.
})
```

#### After

After give you the ability to run a block after each test case.
This is ideal for running cleanups.
The received *testing.T object is the same as the Then block *testing.T object
This hook applied to this scope and anything that is nested from here.
All setup block is stackable.

```go
s := testcase.NewSpec(t)

s.After(func(t *testcase.T) {
    // this will run after the test cases.
    // this hook applied to this scope and anything that is nested from here.
    // hooks can be stacked with each call.
})
```

#### Around

Around give you the ability to create "Before" setup for each test case,
with the additional ability that the returned function will be deferred to run after the Then block is done.
This is ideal for setting up mocks, and then return the assertion request calls in the return func.
This hook applied to this scope and anything that is nested from here.
All setup block is stackable.

```go
s := testcase.NewSpec(t)

s.Around(func(t *testcase.T) func() {
    // this will run before the test cases

    // this hook applied to this scope and anything that is nested from here.
    // hooks can be stacked with each call
    return func() {
        // The content of the returned func will be deferred to run after the test cases.
    }
})
```

### Basic example with Describe+When+Then

documentation:
* [Describe](https://godoc.org/github.com/adamluzsi/testcase#Spec.Describe)
* [When]()
* [Then]()

```go
func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)

    // when no side effect expected,
    // you can use Spec#Parallel for make all test edge case run on different goroutine
    s.Parallel()

    myType := func(t *testcase.T) *MyType {
        return &MyType{Field1: t.I(`input`).(string)}
    }

    s.Describe(`IsLower`, func(s *testcase.Spec) {
        subject := func(t *testcase.T) bool { return myType(t).IsLower() }

        s.When(`input string has lower case characters`, func(s *testcase.Spec) {
            s.Let(`input`, func(t *testcase.T) interface{} { return `all lower case` })

            s.Then(`it will return true`, func(t *testcase.T) {
                t.Parallel()

                if subject(t) != true {
                    t.Fatalf(`it was expected that the %q will re reported to be lowercase`, t.I(`input`))
                }
            })

            s.And(`the first character is capitalized`, func(s *testcase.Spec) {
                s.Let(`input`, func(t *testcase.T) interface{} { return `First character is uppercase` })

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

## The Steps struct based approach

Steps is an easier approach, that allows you to work with vanilla testing pkg T.Run idiom.
It builds on the foundation of variable scoping.
If you use it for setting up variables for your test cases,
you should be aware, that for that purpose, you can only execute your test cases in sequence.

```go
func TestSomething(t *testing.T) {
    var value string

    var steps = testcase.Steps{}
    t.Run(`on`, func(t *testing.T) {
        steps := steps.Before(func(t *testing.T) func() { value = "1"; return func() {} })

        t.Run(`each`, func(t *testing.T) {
            steps := steps.Before(func(t *testing.T) func() { value = "2"; return func() {} })

            t.Run(`nested`, func(t *testing.T) {
                steps := steps.Before(func(t *testing.T) func() { value = "3"; return func() {} })

                t.Run(`layer`, func(t *testing.T) {
                    steps := steps.Before(func(t *testing.T) func() { value = "4"; return func() {} })

                    t.Run(`it will setup and break down the right context`, func(t *testing.T) {
                        steps.Setup(t)

                        require.Equal(t, "4", value)
                    })
                })

                t.Run(`then`, func(t *testing.T) {
                    steps.Setup(t)

                    require.Equal(t, "3", value)
                })
            })

            t.Run(`then`, func(t *testing.T) {
                steps.Setup(t)

                require.Equal(t, "2", value)
            })
        })

        t.Run(`then`, func(t *testing.T) {
            steps.Setup(t)

            require.Equal(t, "1", value)
        })
    })
}
```
