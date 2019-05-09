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

# testcase [GoDoc](https://godoc.org/github.com/adamluzsi/testcase)

The package coverage is 100%

The main documentation is kept in the [GoDoc](https://godoc.org/github.com/adamluzsi/testcase),
and this README serves only as a high level introduction.

This package implements two approaches to help you to do nested BDD style testing in golang.

## [The reason behind the package](https://godoc.org/github.com/adamluzsi/testcase#hdr-The_reason_behind_the_package)
## [What makes testcase different ?](https://godoc.org/github.com/adamluzsi/testcase#hdr-What_makes_testcase_different)

## Reference Project
* [FeatureFlags](https://github.com/adamluzsi/FeatureFlags)

## The Spec based approach

spec structure is a simple wrapping around the testing.T#Run.
It does not use any global singleton cache object or anything like that.
It does not force you to use global variables.

It uses the same idiom as the core go testing pkg also provide you.
You can use the same way as the core testing pkg
> go run ./... -v -run "the/name/of/the/test/it/print/out/in/case/of/failure"

It allows you to do context preparation for each test in a way,
that it will be safe for use with testing.T#Parallel.

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

s.Let(`variable name`, func(v *testcase.V) interface{} {
    return "value"
})

s.Then(`test case`, func(t *testing.T, v *testcase.V) {
    t.Log(v.I(`variable name`).(string)) // -> "value"
})
```

#### Usage within a nested scope

variables strictly belong to a given `Describe`/`When`/`And` scope,
and configured before any hook would be applied,
therefore hooks always receive the most latest version of the `Let` variable,
regardless where they are defined.

```go
func ExampleSpec_Let(t *testing.T) {
	myType := func(v *testcase.V) *MyType { return &MyType{Field1: v.I(`input`).(string)} }

	s := testcase.NewSpec(t)

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(v *testcase.V) bool { return myType(v).IsLower() }

		s.When(`input characters are all lowercase`, func(s *testcase.Spec) {
			s.Let(`input`, func(v *testcase.V) interface{} {
				return "all lowercase"
			})

			s.Then(`it will report true`, func(t *testing.T, v *testcase.V) {
				require.True(t, subject(v))
			})
		})

		s.When(`input is a capitalized`, func(s *testcase.Spec) {
			s.Let(`input`, func(v *testcase.V) interface{} {
				return "Capitalized"
			})

			s.Then(`it will report false`, func(t *testing.T, v *testcase.V) {
				require.False(t, subject(v))
			})
		})
	})
}
```

if your variable can fail, you can use the *V#T function to retrieve the current test run `*testing.T` object.

```go
s.Let(`input`, func(v *testcase.V) interface{} {
	require.True(v.T(), true, `my important test assertion regarding this input variable`)
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

s.Before(func(t *testing.T, v *testcase.V) {
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

s.After(func(t *testing.T, v *testcase.V) {
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

s.Around(func(t *testing.T, v *testcase.V) func() {
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

    myType := func(v *testcase.V) *MyType {
        return &MyType{Field1: v.I(`input`).(string)}
    }

    s.Describe(`IsLower`, func(s *testcase.Spec) {
        subject := func(v *testcase.V) bool { return myType(v).IsLower() }

        s.When(`input string has lower case characters`, func(s *testcase.Spec) {
            s.Let(`input`, func(v *testcase.V) interface{} { return `all lower case` })

            s.Then(`it will return true`, func(t *testing.T, v *testcase.V) {
                t.Parallel()

                if subject(v) != true {
                    t.Fatalf(`it was expected that the %q will re reported to be lowercase`, v.I(`input`))
                }
            })

            s.And(`the first character is capitalized`, func(s *testcase.Spec) {
                s.Let(`input`, func(v *testcase.V) interface{} { return `First character is uppercase` })

                s.Then(`it will report false`, func(t *testing.T, v *testcase.V) {
                    if subject(v) != false {
                        t.Fatalf(`it was expected that %q will be reported to be not lowercase`, v.I(`input`))
                    }
                })
            })
        })
    })
}
```

## The Steps struct based approach

Steps is an easier idiom, that allows you to work with your favorite testing idiom.
It builds on the foundation of variable scoping.
If you use it for setting up variables for your test cases,
you should be aware, that for that purpose, it is not safe to use on concurrent goroutines.

```go
func TestSomething(t *testing.T) {
    var value string

    var steps = testcase.Steps{}
    t.Run(`on`, func(t *testing.T) {
        steps := steps.Add(func(t *testing.T) func() { value = "1"; return func() {} })

        t.Run(`each`, func(t *testing.T) {
            steps := steps.Add(func(t *testing.T) func() { value = "2"; return func() {} })

            t.Run(`nested`, func(t *testing.T) {
                steps := steps.Add(func(t *testing.T) func() { value = "3"; return func() {} })

                t.Run(`layer`, func(t *testing.T) {
                    steps := steps.Add(func(t *testing.T) func() { value = "4"; return func() {} })

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
