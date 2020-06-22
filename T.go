package testcase

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func newT(tb testing.TB) *T {
	t := &T{
		TB:    tb,
		vars:  newVariables(),
		flags: map[string]struct{}{},
	}

	// backward compatibility
	switch e := tb.(type) {
	case *testing.T:
		t.T = e
	}
	return t
}

// T embeds both testcase vars, and testing#T functionality.
// This leave place open for extension and
// but define a stable foundation for the hooks and test edge case function signatures
//
// Works as a drop in replacement for packages where they depend on one of the function of testing#T
//
type T struct {
	T *testing.T
	testing.TB
	vars   *variables
	defers []func()
	flags  map[string]struct{}
}

// I will return a testcase variable.
// it is suggested to use interface casting right after to it,
// so you can work with concrete types.
// If there is no such value, then it will panic with a "friendly" message.
func (t *T) I(varName string) interface{} {
	return t.vars.get(t, varName)
}

// Let will allow you to define/override a spec runtime bounded variable.
// The idiom is that if you cannot express the variable declaration with spec level let,
// or if you need to override in a sub scope a let's content using the previous variable state,
// or a result of a multi return variable needs to be stored at spec runtime level
// you can utilize this Let function to achieve this.
//
// Typical use-case to this when you want to have a context.Context, with different values or states,
// but you don't want to rebuild from scratch at each layer.
func (t *T) Let(varName string, value interface{}) {
	t.vars.set(varName, value)
}

// Defer function defers the execution of a function until the current test case returns.
// Deferred functions are guaranteed to run, regardless of panics during the test case execution.
// Deferred function calls are pushed onto a testcase runtime stack.
// When an function passed to the Defer function, it will be executed as a deferred call in last-in-first-out order.
//
// It is advised to use this inside a testcase.Spec#Let memorization function
// when spec variable defined that has finalizer requirements.
// This allow the specification to ensure the object finalizer requirements to be met,
// without using an testcase.Spec#After where the memorized function would be executed always, regardless of its actual need.
//
// In a practical example, this means that if you have common vars defined with testcase.Spec#Let memorization,
// which needs to be Closed for example, after the test case already run.
// Ensuring such objects Close call in an after block would cause an initialization of the memorized object all the time,
// even in tests where this is not needed.
//
// e.g.:
//	- mock initialization with mock controller, where the mock controller #Finish function must be executed after each test suite.
//	- sql.DB / sql.Tx
//	- basically anything that has the io.Closer interface
//
func (t *T) Defer(fn interface{}, args ...interface{}) {
	rfn := reflect.ValueOf(fn)
	rfnType := rfn.Type()
	if rfn.Kind() != reflect.Func {
		panic(`T#Defer can only take functions`)
	}
	if inCount := rfnType.NumIn(); inCount != len(args) {
		_, file, line, _ := runtime.Caller(1)
		const format = "deferred function argument count mismatch: expected %d, but got %d from %s:%d"
		panic(fmt.Sprintf(format, inCount, len(args), file, line))
	}
	var rargs = make([]reflect.Value, 0, len(args))
	for i, arg := range args {
		value := reflect.ValueOf(arg)
		inType := rfnType.In(i)
		switch expected := inType.Kind(); expected {
		case reflect.Interface:
			if !value.Type().Implements(inType) {
				_, file, line, _ := runtime.Caller(1)
				const format = "deferred function argument[%d] %s doesn't implements %s.%s from %s:%d"
				panic(fmt.Sprintf(format, i, value.Kind(), inType.PkgPath(), inType.Name(), file, line))
			}
		case value.Kind():
			// OK
		default:
			_, file, line, _ := runtime.Caller(1)
			const format = "deferred function argument[%d] type mismatch: expected %s, but got %s from %s:%d"
			panic(fmt.Sprintf(format, i, expected, value.Kind(), file, line))
		}

		rargs = append(rargs, value)
	}
	t.defers = append(t.defers, func() { rfn.Call(rargs) })
}

// In computer science, an operation, function or expression is said to have a side effect if it modifies some state variable value(s) outside its local environment, that is to say has an observable effect besides returning a value (the main effect) to the invoker of the operation. State data updated "outside" of the operation may be maintained "inside" a stateful object or a wider stateful system within which the operation is performed. Example side effects include modifying a non-local variable, modifying a static local variable, modifying a mutable argument passed by reference, performing I/O or calling other side-effect functions.[1] In the presence of side effects, a program's behaviour may depend on history; that is, the order of evaluation matters. Understanding and debugging a function with side effects requires knowledge about the context and its possible histories.[2][3]
// The degree to which side effects are used depends on the programming paradigm. Imperative programming is commonly used to produce side effects, to update a system's state. By contrast, Declarative programming is commonly used to report on the state of system, without side effects.
// In functional programming, side effects are rarely used. The lack of side effects makes it easier to do formal verifications of a program. Functional languages such as Standard ML, Scheme and Scala do not restrict side effects, but it is customary for programmers to avoid them.[4] The functional language Haskell expresses side effects such as I/O and other stateful computations using monadic actions.[5][6]
// Assembly language programmers must be aware of hidden side effects—instructions that modify parts of the processor state which are not mentioned in the instruction's mnemonic. A classic example of a hidden side effect is an arithmetic instruction that implicitly modifies condition codes (a hidden side effect) while it explicitly modifies a register (the overt effect). One potential drawback of an instruction set with hidden side effects is that, if many instructions have side effects on a single piece of state, like condition codes, then the logic required to update that state sequentially may become a performance bottleneck. The problem is particularly acute on some processors designed with pipelining (since 1990) or with out-of-order execution. Such a processor may require additional control circuitry to detect hidden side effects and stall the pipeline if the next instruction depends on the results of those effects.
func (t *T) HasSideEffect() {

}

func (t *T) setup(ctx *context) {
	allCTX := ctx.allLinkListElement()

	t.printDescription(ctx)

	for _, c := range allCTX {
		t.vars.merge(c.vars)
	}

	for _, c := range allCTX {
		for _, hook := range c.hooks {
			t.Defer(hook(t))
		}
	}
}

func (t *T) teardown() {
	for _, td := range t.defers {
		// defer in loop intentionally
		// it will ensure that after hooks are executed
		// at the end of the t.Run block
		// noinspection GoDeferInLoop
		defer td()
	}
}

func (t *T) printDescription(ctx *context) {
	var lines []interface{}

	var spaceIndentLevel int
	for _, c := range ctx.allLinkListElement() {
		if c.description == `` {
			continue
		}

		lines = append(lines, fmt.Sprintln(strings.Repeat(` `, spaceIndentLevel*2), c.description))
		spaceIndentLevel++
	}

	log(t, lines...)
}
