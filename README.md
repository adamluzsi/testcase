# testrun

The package coverage is 100%, and stable.
Changes to the exported signatures are not expected.

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

You receive a new pointer structure called `*testrun.V`
which will represent values that you configured for your test case.
As mentioned above, the values in `*testrun.V` are safe to use during T#Parallel,
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

To me I found it useful, that I always created a `subject`/`asResult` variable with a function that takes `*testrun.V` right after each Spec#Describe function block.
This function signature always shared the same signature as the function/method I test within it.

To me it helped me to have more descriptive test cases, easier refactoring 
and easy way to setup edge cases by using `testrun.Spec#Let`.

On each nesting, I describe the the context about what is the input for example,
or why such case exists, and what is the expected results from it.

This is just a suggest handle it with a grain of salt of course.

### Example

```go
func TestMyStruct(t *testing.T) {

	spec := testrun.NewSpec(t)

	myType := func(v *testrun.V) *MyType {
		return &MyType{Field1: v.I(`input`).(string)}
	}

	spec.Describe(`IsLower`, func(t *testing.T) {
		// it is a convention to me to always make a subject for a certain describe block
		//
		subject := func(v *testrun.V) bool { return myType(v).IsLower() }

		spec.When(`input string has lower case charachers`, func(t *testing.T) {

			spec.Let(`input`, func(v *testrun.V) interface{} {
				return `all lower case`
			})

			spec.Before(func(t *testing.T) {
				// here you can do setups like cleanup for DB tests
			})

			spec.After(func(t *testing.T) {
				// here you can setup teardowns
			})

			spec.Around(func(t *testing.T) func() {
				// here you can setup things that need teardown
				// such example to me is when I use gomock.Controller and mock setup

				return func() {
					// you can do teardown in this
					// this func will be defered after the test cases
				}
			})

			spec.And(`the first character is capitalized`, func(t *testing.T) {
				// you can add more nesting for more concrete specifications,
				// in each nested block, you work on a separate variable stack,
				// so even if you overwrite something here,
				// that has no effect outside of this scope

				spec.Let(`input`, func(v *testrun.V) interface{} {
					return `First character is uppercase`
				})

				spec.Then(`it will report false`, func(t *testing.T, v *testrun.V) {
					if subject(v) != false {
						t.Fatalf(`it was expected that %q will be reported to be not lowercase`, v.I(`input`))
					}
				})

			})

			spec.Then(`it will return true`, func(t *testing.T, v *testrun.V) {
				t.Parallel()

				if subject(v) != true {
					t.Fatalf(`it was expected that the %q will re reported to be lowercase`, v.I(`input`))
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

	var steps = testrun.Steps{}
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