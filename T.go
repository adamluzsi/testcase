package testcase

import (
	"math/rand"
	"testing"
	"time"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/random"

	"github.com/adamluzsi/testcase/internal"
)

// NewT returns a *testcase.T prepared for the given testing.TB
func NewT(tb testing.TB, spec *Spec) *T {
	if spec == nil {
		spec = NewSpec(tb)
	}
	testcaseT := newT(tb, spec)
	tb.Cleanup(testcaseT.setUp())
	return testcaseT
}

func newT(tb testing.TB, spec *Spec) *T {
	return &T{
		TB:     tb,
		Random: random.New(rand.NewSource(spec.seed)),
		Must:   assert.Must(tb),
		Should: assert.Should(tb),

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
	// Random is a random generator that uses the Spec seed.
	//
	// When a test fails with random input from Random generator,
	// the failed test scenario can be recreated simply by providing the same TESTCASE_SEED
	// as you can read from the console output of the failed test.
	Random *random.Random
	// Must Asserter will use FailNow on a failed assertion.
	// This will make test exit early on.
	Must Asserter
	// Should Asserter's will allow to continue the test scenario,
	// but mark test failed on a failed assertion.
	Should Asserter

	spec     *Spec
	vars     *variables
	tags     map[string]struct{}
	teardown *internal.Teardown

	cache struct {
		contexts []*Spec
	}
}

// Asserter contains a minimum set of assertion interactions.
type Asserter interface {
	True(v bool, msg ...interface{})
	False(v bool, msg ...interface{})
	Nil(v interface{}, msg ...interface{})
	NotNil(v interface{}, msg ...interface{})
	Equal(expected, actually interface{}, msg ...interface{})
	NotEqual(expected, actually interface{}, msg ...interface{})
	Contain(source, sub interface{}, msg ...interface{})
	NotContain(source, sub interface{}, msg ...interface{})
	ContainExactly(expected, actually interface{}, msg ...interface{})
	Panic(blk func(), msg ...interface{}) (panicValue interface{})
	NotPanic(blk func(), msg ...interface{})
	Empty(v interface{}, msg ...interface{})
	NotEmpty(v interface{}, msg ...interface{})
}

// I will return a testcase variable.
// it is suggested to use interface casting right after to it,
// so you can work with concrete types.
// If there is no such value, then it will panic with a "friendly" message.
func (t *T) I(varName string) interface{} {
	t.TB.Helper()
	return t.vars.Get(t, varName)
}

// Set will allow you to define/override a spec runtime bounded variable.
// The idiom is that if you cannot express the variable declaration with spec level let,
// or if you need to override in a sub scope a let's content using the previous variable state,
// or a result of a multi return variable needs to be stored at spec runtime level
// you can utilize this Set function to achieve this.
//
// Typical use-case to this when you want to have a spec.Context, with different values or states,
// but you don't want to rebuild from scratch at each layer.
func (t *T) Set(varName string, value interface{}) {
	t.TB.Helper()
	t.vars.Set(varName, value)
}

// Let is an alias for T.Set for backward compatibility.
//
// DEPRECATED: use T.Set instead
func (t *T) Let(varName string, value interface{}) {
	t.TB.Helper()
	t.Logf(`DEPRECATED: testcase.T.Let used to set variable %s value. Consider using testcase.T.Set`, varName)
	t.Set(varName, value)
}

const warnAboutCleanupUsageDuringCleanup = `WARNING: in go1.14 using testing#tb.Cleanup during Cleanup is not supported by the stdlib testing library`

func (t *T) Cleanup(fn func()) {
	t.TB.Helper()
	t.teardown.Defer(fn)
}

// Defer function defers the execution of a function until the current test case returns.
// Deferred functions are guaranteed to run, regardless of panics during the test case execution.
// Deferred function calls are pushed onto a testcase runtime stack.
// When an function passed to the Defer function, it will be executed as a deferred call in last-in-first-orderingOutput order.
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
//	- mock initialization with mock controller, where the mock controller #Finish function must be executed after each testCase suite.
//	- sql.DB / sql.Tx
//	- basically anything that has the io.Closer interface
//
func (t *T) Defer(fn interface{}, args ...interface{}) {
	t.TB.Helper()
	t.teardown.Defer(fn, args...)
}

// setUp resets the *testcase.T cached variable state,
// then set-up all the *testcase.Spec hook and variables in the current *testing.T
// Calling setUp multiple times is safe but it is the caller's responsibility
// to always execute the teardown
func (t *T) setUp() func() {
	t.TB.Helper()
	t.vars.reset()

	contexts := t.contexts()
	for _, c := range contexts {
		t.vars.merge(c.vars)
	}

	for _, c := range contexts {
		for _, hook := range c.hooks.Around {
			t.teardown.Defer(hook(t))
		}
	}

	return t.teardown.Finish
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

func (t *T) hasOnLetHookApplied(name string) bool {
	for _, c := range t.contexts() {
		if ok := c.vars.hasOnLetHookApplied(name); ok {
			return ok
		}
	}
	return false
}

var DefaultEventuallyRetry = Retry{Strategy: Waiter{WaitTimeout: 3 * time.Second}}

// Eventually helper allows you to write expectations to results that will only be eventually true.
// A common scenario where using Eventually will benefit you is testing concurrent operations.
// Due to the nature of async operations, one might need to wait
// and observe the system with multiple tries before the outcome can be seen.
// Eventually will attempt to assert multiple times with the assertion function block,
// until the expectations in the function body yield no testing failure.
// Calling multiple times the assertion function block content should be a safe and repeatable operation.
// For more, read the documentation of Retry and Retry.Assert.
// In case Spec doesn't have a configuration for how to retry Eventually, the DefaultEventuallyRetry will be used.
func (t *T) Eventually(blk func(tb testing.TB)) {
	t.TB.Helper()
	retry, ok := t.spec.lookupRetryEventually()
	if !ok {
		retry = DefaultEventuallyRetry
	}
	retry.Assert(t, blk)
}
