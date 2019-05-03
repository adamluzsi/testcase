# testcase

The package coverage is 100%, and stable.

This package implements two approach to help you to do nested 
"BDD" style testing in golang with testing.T#Run func.

By default if you do nested testing, it will be not BDD of course,
but in my working experience, I found the following idioms kind of productive for creating specifications.

## The Spec based approach

This approach heavily inspirited by the time I spent with working with rspec/jasmine. 

spec structure is a simple wrapping around the testing.T#Run.
It does not use any global singleton cache object, or anything like that.
It does not force you to use global variables.

It use the same idiom as the core go testing pkg also provide you.
You can use the same way as the core testing pkg
> go run ./... -v -run "the/name/of/the/test/it/print/out/in/case/of/failure"
 
It allows you to do context preparation for each test in a way,
that it will be safe for use with testing.T#Parallel.

You receive a new pointer structure called `*testcase.V`
which will represent values that you configured for your test case.
As mentioned above, the values in `*testcase.V` are safe to use during T#Parallel,
so as long your construct does not have any side-effect,
you are free to make run your code on concurrent goroutines.

In case you work with something that depends on side-effects,
such as database tests, you can use the hooks,
to create clean-ahead / clean-up blocks.

Also if you use gomock, you can use the spec#Around function,
to setup the mock with a controller, and in the teardown function,
call the gomock.Controller#Finish function, 
so your test cases will be only about 
what is different behavior from the rest of the test cases.

It will panic if you use hooks or variable preparation in an ambiguous way,
or when you try to access variable that doesn't exist in the context where you do so.
It try to panic with friendly and supportive messages, but that is highly subjective. 

### just a suggestion 

This is here is not really a fancy framework,
it just some basic tooling on top of `*testing.T#Run`.
So it will not give you solutions for everything,
and doesn't even try to do so.

To me I found it useful, that I always created a `subject`/`asResult` variable with a function that takes `*testcase.V` right after each Spec#Describe function block.
This function signature always shared the same signature as the function/method I test within it.

To me it helped me to have more descriptive test cases, easier refactoring 
and easy way to setup edge cases by using `testcase.Spec#Let`.

On each nesting, I describe the the context about what is the input for example,
or why such case exists, and what is the expected results from it.

Then I highly suggest to do blackbox testing,
because then most of the time, your tests can serve as example usages as well.
> to do blackbox testing, just append _test to your current pkg name where you do the testing.

And last but not least, if your implementation need an if,
I suggest to always create two spec context with `When` and `And`
to represent logical decision paths in your specification tree.
Usually it become kind a troublesome to represent too many if,
which helps you realize when your component include too much logic in one place.

But sometimes it is necessary to do so, for those cases,
I usually create a function that takes *testcase.Spec as a receiver,
and do the specification that would otherwise needed to be written redundantly.

This is just a suggest handle it with a grain of salt of course.

### Example

#### Basic example with Describe+When+Then

```go
package mypkg_test

import (
	"github.com/adamluzsi/testcase"
	"strings"
	"testing"
)

type MyType struct {
	Field1 string
}

func (mt *MyType) IsLower() bool {
	return strings.ToLower(mt.Field1) == mt.Field1
}

func TestMyType(t *testing.T) {
	s := testcase.NewSpec(t)

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

#### Complex with hooks 

```go
package mypkg_test

import (
	"github.com/adamluzsi/testcase"
	"strings"
	"testing"
)

type MyType struct {
	Field1 string
}

func (mt *MyType) IsLower() bool {
	return strings.ToLower(mt.Field1) == mt.Field1
}

func (mt *MyType) Fallible() (string, error) {
	return "", nil
}

func ExampleNewSpec(t *testing.T) {

	// spec do not use any global magic
	// it is just a simple abstraction around testing.T#Run
	// Basically you can easily can run it as you would any other go test
	//   -> `go run ./... -v -run "my/edge/case/nested/block/I/want/to/run/only"`
	//
	spec := testcase.NewSpec(t)

	// testcase.V are thread safe way of setting up complex contexts
	// where some variable need to have different values for edge cases.
	// and I usually work with in-memory implementation for certain shared specs,
	// to make my test coverage run fast and still close to somewhat reality in terms of integration.
	// and to me, it is a necessary thing to have "T#Parallel" option safely available
	myType := func(v *testcase.V) *MyType {
		return &MyType{Field1: v.I(`input`).(string)}
	}

	spec.Describe(`IsLower`, func(s *testcase.Spec) {
		// it is a convention to me to always make a subject for a certain describe block
		//
		subject := func(v *testcase.V) bool { return myType(v).IsLower() }

		s.When(`input string has lower case charachers`, func(s *testcase.Spec) {

			s.Let(`input`, func(v *testcase.V) interface{} {
				return `all lower case`
			})

			s.Before(func(t *testing.T) {
				// here you can do setups like cleanup for DB tests
			})

			s.After(func(t *testing.T) {
				// here you can setup teardowns
			})

			s.Around(func(t *testing.T) func() {
				// here you can setup things that need teardown
				// such example to me is when I use gomock.Controller and mock setup

				return func() {
					// you can do teardown in this
					// this func will be defered after the test cases
				}
			})

			s.And(`the first character is capitalized`, func(s *testcase.Spec) {
				// you can add more nesting for more concrete specifications,
				// in each nested block, you work on a separate variable stack,
				// so even if you overwrite something here,
				// that has no effect outside of this scope

				s.Let(`input`, func(v *testcase.V) interface{} {
					return `First character is uppercase`
				})

				s.Then(`it will report false`, func(t *testing.T, v *testcase.V) {
					if subject(v) != false {
						t.Fatalf(`it was expected that %q will be reported to be not lowercase`, v.I(`input`))
					}
				})

			})

			s.Then(`it will return true`, func(t *testing.T, v *testcase.V) {
				t.Parallel()

				if subject(v) != true {
					t.Fatalf(`it was expected that the %q will re reported to be lowercase`, v.I(`input`))
				}
			})
		})
	})

	spec.Describe(`Fallible`, func(s *testcase.Spec) {

		subject := func(v *testcase.V) (string, error) {
			return myType(v).Fallible()
		}

		onSuccessfulRun := func(t *testing.T, v *testcase.V) string {
			someMeaningfulVarName, err := subject(v)
			if err != nil {
				t.Fatal(err.Error())
			}
			return someMeaningfulVarName
		}

		s.When(`input is an empty string`, func(s *testcase.Spec) {
			s.Let(`input`, func(v *testcase.V) interface{} { return "" })

			s.Then(`it will return an empty string`, func(t *testing.T, v *testcase.V) {
				if res := onSuccessfulRun(t, v); res != "" {
					t.Fatalf(`it should have been an empty string, but it was %q`, res)
				}
			})

		})

	})
}
```

## Steps based approach

Steps is a easier idiom, that allow you to work with your favorite testing idiom.
It builds on the foundation of variable scoping.
If you use it for setting up variables for your test cases,
you should be aware, that for that purpose, it is not safe to use on concurrent goroutines.

```go
func TestSomething(t *testing.T) {
	var value string

	var steps = testcase.Steps{}
	t.Run(`on`, func(t *testing.T) {
		steps := steps.Add(func(t *testing.T) { value = "1" })

		t.Run(`each`, func(t *testing.T) {
			steps := steps.Add(func(t *testing.T) { value = "2" })

			t.Run(`nested`, func(t *testing.T) {
				steps := steps.Add(func(t *testing.T) { value = "3" })

				t.Run(`layer`, func(t *testing.T) {
					steps := steps.Add(func(t *testing.T) { value = "4" })

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

## Reference Projects
* [FeatureFlags](https://github.com/adamluzsi/FeatureFlags)
  * root cause why I created this in the first place.
