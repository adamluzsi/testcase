package testcase

import (
	"math/rand"
	"testing"
	"time"

	"github.com/adamluzsi/testcase/pp"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/teardown"
	"github.com/adamluzsi/testcase/random"
)

// NewT returns a *testcase.T prepared for the given testing.TB
func NewT(tb testing.TB, spec *Spec) *T {
	if tb == nil {
		return nil
	}
	if tcT, ok := tb.(*T); ok {
		return tcT
	}
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
		Random: random.New(rand.NewSource(spec.getTestSeed(tb))),
		It:     assert.MakeIt(tb),

		spec:     spec,
		vars:     newVariables(),
		tags:     spec.getTagSet(),
		teardown: &teardown.Teardown{CallerOffset: 1},
	}
}

// T embeds both testcase vars, and testing#T functionality.
// This leave place open for extension and
// but define a stable foundation for the hooks and testCase edge case function signatures
//
// Works as a drop in replacement for packages where they depend on one of the function of testing#T
type T struct {
	// TB is the interface common to T and B.
	testing.TB
	// Random is a random generator that uses the Spec seed.
	//
	// When a test fails with random input from Random generator,
	// the failed test scenario can be recreated simply by providing the same TESTCASE_SEED
	// as you can read from the console output of the failed test.
	Random *random.Random
	// It provides asserters to make assertion easier.
	// Must Interface will use FailNow on a failed assertion.
	// This will make test exit early on.
	// Should Interface's will allow to continue the test scenario,
	// but mark test failed on a failed assertion.
	assert.It

	spec     *Spec
	vars     *variables
	tags     map[string]struct{}
	teardown *teardown.Teardown

	// TODO: protect it against concurrency
	timerPaused bool

	cache struct {
		contexts []*Spec
	}
}

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
//   - mock initialization with mock controller, where the mock controller #Finish function must be executed after each testCase suite.
//   - sql.DB / sql.Tx
//   - basically anything that has the io.Closer interface
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
		for _, hook := range c.hooks.BeforeAll {
			hook.Block()
		}
	}

	for _, c := range contexts {
		for _, hook := range c.hooks.Around {
			t.teardown.Defer(hook.Block(t))
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
		t.cache.contexts = t.spec.specsFromParent()
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

var DefaultEventually = assert.Eventually{RetryStrategy: assert.Waiter{Timeout: 3 * time.Second}}

// Eventually helper allows you to write expectations to results that will only be eventually true.
// A common scenario where using Eventually will benefit you is testing concurrent operations.
// Due to the nature of async operations, one might need to wait
// and observe the system with multiple tries before the outcome can be seen.
// Eventually will attempt to assert multiple times with the assertion function block,
// until the expectations in the function body yield no testing failure.
// Calling multiple times the assertion function block content should be a safe and repeatable operation.
// For more, read the documentation of Eventually and Eventually.Assert.
// In case Spec doesn't have a configuration for how to retry Eventually, the DefaultEventually will be used.
func (t *T) Eventually(blk func(t assert.It), retryOpts ...interface{}) {
	t.TB.Helper()
	retry, ok := t.spec.lookupRetryEventually()
	if !ok {
		retry = DefaultEventually
	}
	retry.Assert(t, blk)
}

type timerManager interface {
	StartTimer()
	StopTimer()
	ResetTimer()
}

func (t *T) pauseTimer() func() {
	t.TB.Helper()
	btm, ok := t.TB.(timerManager)
	if !ok {
		return func() {}
	}
	if t.timerPaused {
		return func() {}
	}

	btm.StopTimer()
	t.timerPaused = true
	return func() {
		t.timerPaused = false
		btm.StartTimer()
	}
}

// SkipUntil is equivalent to SkipNow if the test is executing prior to the given deadline time.
// SkipUntil is useful when you need to skip something temporarily, but you don't trust your memory enough to return to it on your own.
func (t *T) SkipUntil(year int, month time.Month, day int, hour int) {
	t.TB.Helper()
	SkipUntil(t.TB, year, month, day, hour)
}

const envMutationDuringParallelExecution = "Env variables manipulated during Parallel test execution, please use Spec.HasSideEffect or Spec.Sequential"

// UnsetEnv will unset the os environment variable value for the current program,
// and prepares a cleanup function to restore the original state of the environment variable.
//
// This cannot be used in parallel tests.
func (t *T) UnsetEnv(key string) {
	t.Helper()
	if t.spec.isParallel() {
		t.Fatal(envMutationDuringParallelExecution)
	}
	UnsetEnv(t, key)
}

// SetEnv will set the os environment variable for the current program to a given value,
// and prepares a cleanup function to restore the original state of the environment variable.
//
// This cannot be used in parallel tests.
func (t *T) SetEnv(key, value string) {
	t.Helper()
	if t.spec.isParallel() {
		t.Fatal(envMutationDuringParallelExecution)
	}
	SetEnv(t.TB, key, value)
}

// Setenv calls os.Setenv(key, value) and uses Cleanup to
// restore the environment variable to its original value
// after the test.
//
// This cannot be used in parallel tests.
func (t *T) Setenv(key, value string) {
	t.Helper()
	t.SetEnv(key, value)
}

// LogPretty will Log out values in pretty print format (pp.Format).
func (t *T) LogPretty(vs ...any) {
	t.Helper()
	var args []any
	for _, v := range vs {
		args = append(args, pp.Format(v))
	}
	t.Log(args...)
}
