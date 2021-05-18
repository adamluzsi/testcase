package testcase

import (
	"math/rand"
	"testing"

	"github.com/adamluzsi/testcase/random"

	"github.com/adamluzsi/testcase/internal"
)

func newT(tb testing.TB, spec *Spec) *T {
	tb.Helper()
	return &T{
		TB:     tb,
		Random: random.New(rand.NewSource(spec.seed)),

		spec:     spec,
		vars:     newVariables(),
		tags:     spec.getTagSet(),
		teardown: &internal.Teardown{CallerOffset: 1},
	}
}

// T embeds both testcase vars, and testing#T functionality.
// This leave place open for extension and
// but define a stable foundation for the hooks and testCase edge case function signatures
//
// Works as a drop in replacement for packages where they depend on one of the function of testing#T
//
type T struct {
	// TB is the interface common to T and B.
	testing.TB
	// Random is a random generator that use the Spec seed.
	//
	// When a test fails with a random input, the failure can be recreated simply by providing the same TESTCASE_SEED.
	Random *random.Random

	spec     *Spec
	vars     *variables
	tags     map[string]struct{}
	teardown *internal.Teardown

	cache struct {
		contexts []*Spec
	}
}

// I will return a testcase variable.
// it is suggested to use interface casting right after to it,
// so you can work with concrete types.
// If there is no such value, then it will panic with a "friendly" message.
func (t *T) I(varName string) interface{} {
	t.TB.Helper()
	return t.vars.Get(t, varName)
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
	t.TB.Helper()
	t.vars.Set(varName, value)
}

const warnAboutCleanupUsageDuringCleanup = `WARNING: in go1.14 using testing#tb.Cleanup during Cleanup is not supported by the stdlib testing library`

func (t *T) Cleanup(fn func()) {
	t.TB.Helper()
	t.teardown.Defer(fn)
}

// Defer function defers the execution of a function until the current testCase case returns.
// Deferred functions are guaranteed to run, regardless of panics during the testCase case execution.
// Deferred function calls are pushed onto a testcase runtime stack.
// When an function passed to the Defer function, it will be executed as a deferred call in last-in-first-orderingOutput order.
//
// It is advised to use this inside a testcase.Spec#Let memorization function
// when spec variable defined that has finalizer requirements.
// This allow the specification to ensure the object finalizer requirements to be met,
// without using an testcase.Spec#After where the memorized function would be executed always, regardless of its actual need.
//
// In a practical example, this means that if you have common vars defined with testcase.Spec#Let memorization,
// which needs to be Closed for example, after the testCase case already run.
// Ensuring such objects Close call in an after block would cause an initialization of the memorized object list the time,
// even in tests where this is not needed.
//
// e.g.:
//	- mock initialization with mock controller, where the mock controller #Finish function must be executed after each testCase suite.
//	- sql.DB / sql.Tx
//	- basically anything that has the io.Closer interface
//
func (t *T) Defer(fn interface{}, args ...interface{}) {
	t.TB.Helper()
	t.teardown.Defer(fn, args...)
}

func (t *T) HasTag(tag string) bool {
	t.TB.Helper()
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
	t.TB.Helper()
	t.vars.reset()

	contexts := t.contexts()
	for _, c := range contexts {
		t.vars.merge(c.vars)
	}

	for _, c := range contexts {
		for _, hook := range c.hooks {
			t.teardown.Defer(hook(t))
		}
	}

	return t.teardown.Finish
}

func (t *T) hasOnLetHookApplied(name string) bool {
	for _, c := range t.contexts() {
		if ok := c.vars.hasOnLetHookApplied(name); ok {
			return ok
		}
	}
	return false
}
