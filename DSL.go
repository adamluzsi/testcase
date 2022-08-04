package testcase

import (
	"fmt"
)

// Describe creates a new spec scope, where you usually describe a subject.
//
// By convention it is highly advised to create a variable `subject`
// with function that share the return signature of the method you test on a structure,
// and take *testcase.variables as the only input value.
// If your method require input values, you should strictly set those values within a `When`/`And` scope.
// This ensures you have to think trough the possible state-machines paths that are based on the input values.
//
// For functions where 2 value is returned, and the second one is an error,
// in order to avoid repetitive test cases in the `Then` I often define a `onSuccess` variable,
// with a function that takes `testcase#variables` as well and test error return value there with `testcase#variables.T()`.
//
func (spec *Spec) Describe(subjectTopic string, specification sBlock, opts ...SpecOption) {
	spec.testingTB.Helper()
	opts = append([]SpecOption{Group(subjectTopic)}, opts...)
	spec.Context(fmt.Sprintf(`%s %s`, `describe`, subjectTopic), specification, opts...)
}

// When is an alias for testcase#Spec.Context
// When is used usually to represent `if` based decision reasons about your testing subject.
func (spec *Spec) When(desc string, testContextBlock sBlock, opts ...SpecOption) {
	spec.testingTB.Helper()
	spec.Context(fmt.Sprintf(`%s %s`, `when`, desc), testContextBlock, opts...)
}

// And is an alias for testcase#Spec.Context
// And is used to represent additional requirement for reaching a certain testing runtime contexts.
func (spec *Spec) And(desc string, testContextBlock sBlock, opts ...SpecOption) {
	spec.testingTB.Helper()
	spec.Context(fmt.Sprintf(`%s %s`, `and`, desc), testContextBlock, opts...)
}

// Then is an alias for Test
func (spec *Spec) Then(desc string, test tBlock, opts ...SpecOption) {
	spec.testingTB.Helper()
	desc = fmt.Sprintf(`%s %s`, `then`, desc)
	spec.Test(desc, test, opts...)
}

// NoSideEffect gives a hint to the reader of the current test that during the test execution,
// no side effect outside from the test specification scope is expected to be observable.
// It is important to note that this flag primary meant to represent the side effect possibility
// to the outside of the current testing specification,
// and not about the test specification's subject.
//
// It is safe to state that if the subject of the test specification has no side effect,
// then the test specification must have no side effect as well.
//
// If the subject of the test specification do side effect on an input value,
// then the test specification must have no side effect, as long Let memorization is used.
//
// If the subject of the test specification does mutation on global variables
// such as OS Variable states for the current process,
// then it is likely, that even if the changes by the mutation is restored as part of the test specification,
// the test specification has side effects that would affect other test specification results,
// and, as such, must be executed sequentially.
func (spec *Spec) NoSideEffect() {
	spec.testingTB.Helper()
	spec.Parallel()
}

// HasSideEffect means that after this call things defined that has software side effect during runtime.
// This suggest on its own that execution should be sequential in order to avoid retry tests.
//
// HasSideEffect and NoSideEffect can be used together to describe a given piece of specification properties.
// Using them at the same location makes little sense,
// it was intended to be used in spec helper package where setup function handles what resource should be used in the spec variables.
// This allows flexibility for the developers to use side effect free variant for local development that has quick feedback loop,
// and replace them with the production implementation during CI/CD pipeline which less time critical.
func (spec *Spec) HasSideEffect() {
	spec.testingTB.Helper()
	spec.Sequential()
}

// PreviousValue is a syntax sugar for Var.Super to access the previous declaration's value.
func (v Var[V]) PreviousValue(t *T) V {
	t.Helper()
	return v.Super(t)
}
