package testrun_test

import (
	"github.com/adamluzsi/testrun"
	"strings"
	"testing"
)

type MyType struct {
	Field1 string
}

func (mt *MyType) IsLower() bool {
	return strings.ToLower(mt.Field1) == mt.Field1
}

func ExampleNewSpec(t *testing.T) {

	// spec do not use any global magic
	// it is just a simple abstraction around testing.T#Run
	// Basically you can easily can run it as you would any other go test
	//   -> `go run ./... -v -run "my/edge/case/nested/block/I/want/to/run/only"`
	//
	spec := testrun.NewSpec(t)

	// testrun.V are thread safe way of setting up complex contexts
	// where some variable need to have different values for edge cases.
	// and I usually work with in-memory implementation for certain shared specs,
	// to make my test coverage run fast and still close to somewhat reality in terms of integration.
	// and to me, it is a necessary thing to have "T#Parallel" option safely available
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
