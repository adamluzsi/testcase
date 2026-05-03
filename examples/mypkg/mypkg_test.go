package mypkg_test

import (
	"strings"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/examples/mypkg"
	"go.llib.dev/testcase/let"
)

// TestNameOfSystemUnderTest
//
// The name should always be the testing subject's name.
// If it is a Type (e.g.: MyType), then TestType (e.g.: TestMyType)
// Sub testing context will be expressed with testcase.Spec#Context function blocks.
func TestNameOfSystemUnderTest(t *testing.T) {
	s := testcase.NewSpec(t)

	subject := let.Var(s, func(t *testcase.T) mypkg.MyType {
		return mypkg.MyType{}
	})

	// The describe block used to describe the mypkg.MyType#MyFunc method.
	// As a rule of thumb, the Describe block MUST have a dedicated ACT defined in the top,
	// that the describe block itself meant to describe.
	//
	// The ACT might have its own inputs arguments, if the ACT requires input arguments.
	s.Describe("#MyFunc", func(s *testcase.Spec) {
		// var block with () helps to visualise nicely the input parameters of the immutable testing ACT
		var (
			// input is a testcase variable
			input = let.Var(s, func(t *testcase.T) string {
				// t.Random is a pseudo deterministic random input.
				// It helps with incorporating property testing into your test.
				// Unless you expect a concrete input type, you can use various pseudo randoms with t.Random,
				// and if the test breaks due to unhandled input value from the random,
				// you will get back the TESTCASE_SEED number that can recreate the failing testing scenario 1:1.
				return t.Random.String()
			})
		)
		// act is the immutable testing ACT of this describe block.
		// The importance of this, that this practice forces you to arrange the inputs and the context of your tests
		// instead of using the act function variously within tests.
		// It also reduces the mental model to understand how the ACT will happen within the testing scenarios,
		// and help to shift the focus to the context building
		act := func(t *testcase.T) string {
			return subject.Get(t).MyFunc(input.Get(t))
		}

		// first and most simpistic happy path's behaviour aimed at the top level of the Describe block.
		// We should avoid the need to have deep context nesting to reach a happy path, that should be the default target.
		s.Then("input returned as is", func(t *testcase.T) {
			assert.Equal(t, act(t), input.Get(t))
		})

		// when something is not the most simplistic happy path behaviour,
		// they expressed with a Context/When/And context building blocks.
		// As a rule of thumb, Context/When/And MUST start with something that ARRANGE what the context describes.
		//
		// It could be a modification of an existing FACT of the test, for example, a variable value changed within that context with testcase.Var#Let
		// or with testcase.Spec#Before hook.
		//
		// Keep in mind that redefining a variable will ensures that the variable's value will be changed from the get-go,
		// while the Spec#Before hooks applied sequentially, starting from the outer context hooks, towards inwards towards where the test is located.
		//
		// It also helps with keeping your test arrangements DRY, because any new test or sub context with a context will inherit the arrangement.
		s.When("MyType#ToUpper option is set", func(s *testcase.Spec) {
			subject.Let(s, func(t *testcase.T) mypkg.MyType {
				// we get the previous / super value of the subject testcase.Var
				sub := subject.Super(t)
				// we modify it to fullfil the context's goal
				sub.ToUpper = true
				// and return the newly composed value that achieves this
				return sub
			})

			s.Then("the result will be in upper format", func(t *testcase.T) {
				assert.Equal(t, act(t), strings.ToUpper(input.Get(t)))
			})
		})
	})
}
