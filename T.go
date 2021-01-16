package testcase

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func newT(tb testing.TB, spec *Spec) *T {
	return &T{
		TB:   tb,
		spec: spec,
		vars: newVariables(),
		tags: spec.getTagSet(),
	}
}

// T embeds both testcase vars, and testing#T functionality.
// This leave place open for extension and
// but define a stable foundation for the hooks and test edge case function signatures
//
// Works as a drop in replacement for packages where they depend on one of the function of testing#T
//
type T struct {
	testing.TB

	spec     *Spec
	vars     *variables
	tags     map[string]struct{}
	cleanups []func()

	cache struct {
		contexts []*Spec
	}
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
// Typical use-case to this when you want to have a spec.Context, with different values or states,
// but you don't want to rebuild from scratch at each layer.
func (t *T) Let(varName string, value interface{}) {
	t.vars.set(varName, value)
}

func (t *T) Cleanup(fn func()) {
	t.cleanups = append(t.cleanups, fn)
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
// Ensuring such objects Close call in an after block would cause an initialization of the memorized object list the time,
// even in tests where this is not needed.
//
// e.g.:
//	- mock initialization with mock controller, where the mock controller #Finish function must be executed after each test suite.
//	- sql.DB / sql.Tx
//	- basically anything that has the io.Closer interface
//
func (t *T) Defer(fn interface{}, args ...interface{}) {
	if fn, ok := fn.(func()); ok && len(args) == 0 {
		t.Cleanup(fn)
		return
	}

	rfn := reflect.ValueOf(fn)
	if rfn.Kind() != reflect.Func {
		panic(`T#Defer can only take functions`)
	}
	rfnType := rfn.Type()
	if inCount := rfnType.NumIn(); inCount != len(args) {
		_, file, line, _ := runtime.Caller(1)
		const format = "deferred function argument count mismatch: expected %d, but got %d from %s:%d"
		panic(fmt.Sprintf(format, inCount, len(args), file, line))
	}
	var refArgs = make([]reflect.Value, 0, len(args))
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

		refArgs = append(refArgs, value)
	}

	t.Cleanup(func() { rfn.Call(refArgs) })
}

func (t *T) HasTag(tag string) bool {
	_, ok := t.tags[tag]
	return ok
}

func (t *T) contexts() []*Spec {
	if t.cache.contexts == nil {
		t.cache.contexts = t.spec.list()
	}
	return t.cache.contexts
}

func (t *T) setup() func() {
	contexts := t.contexts()

	t.cleanups = nil
	t.vars.reset()
	for _, c := range contexts {
		t.vars.merge(c.vars)
	}

	for _, c := range contexts {
		for _, hook := range c.hooks {
			t.Cleanup(hook(t))
		}
	}

	return func() {
		for _, td := range t.cleanups {
			switch t.TB.(type) {
			case *testing.B:
				defer td()
			default:
				t.TB.Cleanup(td)
			}
		}
	}
}
