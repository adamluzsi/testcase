package testcase_test

import (
	"strings"
	"testing"

	"github.com/adamluzsi/testcase"
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

	// when you have no side effects in your testing suite,
	// you can enable Parallel execution.
	// You can Call Parallel even from nested specs to apply Parallel testing for that context and below.
	spec.Parallel()

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

		s.When(`input string has lower case characters`, func(s *testcase.Spec) {

			s.Let(`input`, func(v *testcase.V) interface{} {
				return `all lower case`
			})

			s.Before(func(t *testing.T, v *testcase.V) {
				// here you can do setups like cleanup for DB tests
			})

			s.After(func(t *testing.T, v *testcase.V) {
				// here you can setup a teardown
			})

			s.Around(func(t *testing.T, v *testcase.V) func() {
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
