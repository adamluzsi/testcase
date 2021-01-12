package testcase

import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

// NewSpec create new Spec struct that is ready for usage.
func NewSpec(tb testing.TB) *Spec {
	return &Spec{
		testingTB: tb,
		context:   newContext(nil),
	}
}

func (spec *Spec) newSubSpec(desc string) *Spec {
	spec.context.immutable = true
	subCTX := newContext(spec.context)
	subCTX.description = desc
	return &Spec{
		testingTB: spec.testingTB,
		context:   subCTX,
	}
}

// Spec provides you a struct that makes building nested test context easy with the core T#Context function.
//
// spec structure is a simple wrapping around the testing.T#Context.
// It doesn't use any global singleton cache object or anything like that.
// It doesn't force you to use global vars.
//
// It uses the same idiom as the core go testing pkg also provide you.
// You can use the same way as the core testing pkg
// 	go run ./... -v -run "the/name/of/the/test/it/print/out/in/case/of/failure"
//
// It allows you to do context preparation for each test in a way,
// that it will be safe for use with testing.T#Parallel.
type Spec struct {
	testingTB testing.TB
	context   *context
}

// Context allow you to create a sub specification for a given spec.
// In the sub-specification it is expected to add more contextual information to the test
// in a form of hook of variable setting.
// With Context you can set your custom test description, without any forced prefix like describe/when/and.
//
// It is basically piggybacking the testing#T.Context and create new subspec in that nested testing#T.Context scope.
// It is used to add more description context for the given subject.
// It is highly advised to always use When + Before/Around together,
// in which you should setup exactly what you wrote in the When description input.
// You can Context as many When/And within each other, as you want to achieve
// the most concrete edge case you want to test.
//
// To verify easily your state-machine, you can count the `if`s in your implementation,
// and check that each `if` has 2 `When` block to represent the two possible path.
//
func (spec *Spec) Context(desc string, testContextBlock func(s *Spec), opts ...ContextOption) {
	s := spec.newSubSpec(desc)
	for _, opt := range opts {
		opt.setup(s.context)
	}

	if s.context.name == `` {
		testContextBlock(s)
		return
	}

	run := func(tb testing.TB, blk func(*Spec)) {
		switch tb := tb.(type) {
		case *testing.T:
			tb.Run(s.context.name, func(t *testing.T) {
				s.testingTB = t
				blk(s)
			})

		case *testing.B:
			tb.Run(s.context.name, func(b *testing.B) {
				s.testingTB = b
				blk(s)
			})

		default:
			blk(s)
		}
	}

	run(spec.testingTB, testContextBlock)
}

type testCaseBlock func(*T)

// Test creates a test case block where you receive the fully configured `testcase#T` object.
// Hook contents that meant to run before the test edge cases will run before the function the Test receives,
// and hook contents that meant to run after the test edge cases will run after the function is done.
// After hooks are deferred after the received function block, so even in case of panic, it will still be executed.
//
// It should not contain anything that modify the test subject input.
// It should focuses only on asserting the result of the subject.
//
func (spec *Spec) Test(desc string, test testCaseBlock, opts ...ContextOption) {
	s := spec.newSubSpec(desc)
	for _, to := range opts {
		to.setup(s.context)
	}
	s.run(test)
}

// Before give you the ability to run a block before each test case.
// This is ideal for doing clean ahead before each test case.
// The received *testing.T object is the same as the Test block *testing.T object
// This hook applied to this scope and anything that is nested from here.
// All setup block is stackable.
func (spec *Spec) Before(beforeBlock testCaseBlock) {
	spec.context.addHook(func(t *T) func() {
		beforeBlock(t)
		return func() {}
	})
}

// After give you the ability to run a block after each test case.
// This is ideal for running cleanups.
// The received *testing.T object is the same as the Then block *testing.T object
// This hook applied to this scope and anything that is nested from here.
// All setup block is stackable.
func (spec *Spec) After(afterBlock testCaseBlock) {
	spec.context.addHook(func(t *T) func() {
		return func() { afterBlock(t) }
	})
}

type hookBlock func(*T) func()

// Around give you the ability to create "Before" setup for each test case,
// with the additional ability that the returned function will be deferred to run after the Then block is done.
// This is ideal for setting up mocks, and then return the assertion request calls in the return func.
// This hook applied to this scope and anything that is nested from here.
// All setup block is stackable.
func (spec *Spec) Around(aroundBlock hookBlock) {
	spec.context.addHook(aroundBlock)
}

const warnEventOnImmutableFormat = `you can't use #%s after you already used when/and/then`

// Parallel allows you to set all test case for the context where this is being called,
// and below to nested contexts, to be executed in parallel (concurrently).
// Keep in mind that you can call Parallel even from nested specs
// to apply Parallel testing for that context and below.
// This is useful when your test suite has no side effects at all.
// Using values from *vars when Parallel is safe.
// It is a shortcut for executing *testing.T#Parallel() for each test
func (spec *Spec) Parallel() {
	if spec.context.immutable {
		panic(fmt.Sprintf(warnEventOnImmutableFormat, `Parallel`))
	}

	parallel().setup(spec.context)
}

// SkipBenchmark will flag the current Spec / Context to be skipped during Benchmark mode execution.
// If you wish to skip only a certain test, not the whole Spec / Context, use the SkipBenchmark ContextOption instead.
func (spec *Spec) SkipBenchmark() {
	if spec.context.immutable {
		panic(fmt.Sprintf(warnEventOnImmutableFormat, `SkipBenchmark`))
	}

	SkipBenchmark().setup(spec.context)
}

// Sequential allows you to set all test case for the context where this is being called,
// and below to nested contexts, to be executed sequentially.
// It will negate any testcase.Spec#Parallel call effect.
// This is useful when you want to create a spec helper package
// and there you want to manage if you want to use components side effects or not.
func (spec *Spec) Sequential() {
	if spec.context.immutable {
		panic(fmt.Sprintf(warnEventOnImmutableFormat, `Sequential`))
	}

	sequential().setup(spec.context)
}

// Skip is equivalent to Log followed by SkipNow on T for each test case.
func (spec *Spec) Skip(args ...interface{}) {
	spec.Before(func(t *T) { t.TB.Skip(args...) })
}

// Let define a memoized helper method.
// Let creates lazily-evaluated test execution bound variables.
// Let variables don't exist until called into existence by the actual tests,
// so you won't waste time loading them for examples that don't use them.
// They're also memoized, so they're useful for encapsulating database objects, due to the cost of making a database request.
// The value will be cached across all use within the same test execution but not across different test cases.
// You can eager load a value defined in let by referencing to it in a Before hook.
// Let is threadsafe, the parallel running test will receive they own test variable instance.
//
// Defining a value in a spec Context will ensure that the scope
// and it's nested scopes of the current scope will have access to the value.
// It cannot leak its value outside from the current scope.
// Calling Let in a nested/sub scope will apply the new value for that value to that scope and below.
//
// It will panic if it is used after a When/And/Then scope definition,
// because those scopes would have no clue about the later defined variable.
// In order to keep the specification reading mental model requirement low,
// it is intentionally not implemented to handle such case.
// Defining test vars always expected in the beginning of a specification scope,
// mainly for readability reasons.
//
// vars strictly belong to a given `Describe`/`When`/`And` scope,
// and configured before any hook would be applied,
// therefore hooks always receive the most latest version from the `Let` vars,
// regardless in which scope the hook that use the variable is define.
//
// Let can enhance readability
// when used sparingly in any given example group,
// but that can quickly degrade with heavy overuse.
//
func (spec *Spec) Let(varName string, blk letBlock) Var {
	if spec.context.immutable {
		panic(fmt.Sprintf(warnEventOnImmutableFormat, `Let/LetValue`))
	}

	spec.context.let(varName, blk)

	return Var{Name: varName, Init: blk}
}

type letBlock func(t *T) /* T */ interface{}

var acceptedConstKind = map[reflect.Kind]struct{}{
	reflect.String:     {},
	reflect.Bool:       {},
	reflect.Int:        {},
	reflect.Int8:       {},
	reflect.Int16:      {},
	reflect.Int32:      {},
	reflect.Int64:      {},
	reflect.Uint:       {},
	reflect.Uint8:      {},
	reflect.Uint16:     {},
	reflect.Uint32:     {},
	reflect.Uint64:     {},
	reflect.Float32:    {},
	reflect.Float64:    {},
	reflect.Complex64:  {},
	reflect.Complex128: {},
}

const panicMessageForLetValue = `%T literal can't be used with #LetValue 
as the current implementation can't guarantee that the mutations on the value will not leak out to other tests,
please use the #Let memorization helper for now`

// LetValue is a shorthand for defining immutable vars with Let under the hood.
// So the function blocks can be skipped, which makes tests more readable.
func (spec *Spec) LetValue(varName string, value interface{}) Var {
	if _, ok := acceptedConstKind[reflect.ValueOf(value).Kind()]; !ok {
		panic(fmt.Sprintf(panicMessageForLetValue, value))
	}

	return spec.Let(varName, func(t *T) interface{} {
		v := value // pass by value copy
		return v
	})
}

// Tag allow you to mark tests in the current and below specification scope with tags.
// This can be used to provide additional documentation about the nature of the testing scope.
// This later might be used as well to filter your test in your CI/CD pipeline to build separate testing stages like integration, e2e and so on.
//
// To select or exclude tests with certain tags, you can provide a comma separated list to the following environment variables:
//  - TESTCASE_TAG_INCLUDE to filter down to test with a certain tag
//  - TESTCASE_TAG_EXCLUDE to exclude certain test from the overall testing scope.
// They can be combined as well.
//
// example usage:
// 	TESTCASE_TAG_INCLUDE='E2E' go test ./...
// 	TESTCASE_TAG_EXCLUDE='E2E' go test ./...
// 	TESTCASE_TAG_INCLUDE='E2E' TESTCASE_TAG_EXCLUDE='list,of,excluded,tags' go test ./...
//
func (spec *Spec) Tag(tags ...string) {
	spec.context.tags = append(spec.context.tags, tags...)
}

func (spec *Spec) isAllowedToRun() bool {
	currentTagSet := spec.context.getTagSet()
	settings := getCachedTagSettings()

	for tag := range currentTagSet {
		if _, ok := settings.Exclude[tag]; ok {
			return false
		}
	}

	if len(settings.Include) == 0 {
		return true
	}

	var allowed bool
	for tag := range currentTagSet {
		if _, ok := settings.Include[tag]; ok {
			allowed = true
		}
	}
	return allowed
}

func (spec *Spec) isBenchAllowedToRun() bool {
	for _, context := range spec.context.all() {
		if context.skipBenchmark {
			return false
		}
	}
	return true
}

func (spec *Spec) lookupFlaky() (*flakyFlag, bool) {
	for _, context := range spec.context.all() {
		if context.flaky != nil {
			return context.flaky, true
		}
	}
	return nil, false
}

func (spec *Spec) printDescription(t *T) {
	var lines []interface{}

	var spaceIndentLevel int
	for _, c := range t.contexts() {
		if c.description == `` {
			continue
		}

		lines = append(lines, fmt.Sprintln(strings.Repeat(` `, spaceIndentLevel*2), c.description))
		spaceIndentLevel++
	}

	log(t, lines...)
}

func (spec *Spec) name() string {
	switch spec.testingTB.(type) {
	case *testing.B:
		var desc string
		for _, context := range spec.context.all() {
			if desc != `` {
				desc += ` `
			}

			desc += context.description
		}
		return desc

	default:
		return ``
	}
}

///////////////////////////////////////////////////////=- run -=////////////////////////////////////////////////////////

func (spec *Spec) run(blk func(*T)) {
	if !spec.isAllowedToRun() {
		return
	}

	switch tb := spec.testingTB.(type) {
	case *testing.T:
		tb.Run(spec.name(), func(t *testing.T) {
			spec.runTB(t, blk)
		})
	case *testing.B:
		if !spec.isBenchAllowedToRun() {
			return
		}
		tb.Run(spec.name(), func(b *testing.B) {
			spec.runB(b, blk)
		})
	case CustomTB:
		tb.Run(spec.name(), func(tb testing.TB) {
			spec.runTB(tb, blk)
		})

	default:
		panic(fmt.Errorf(`test runner %T is unsupported, please implement testcase.CustomTB`, tb))
	}
}

func (spec *Spec) runTB(tb testing.TB, blk func(*T)) {
	if tb, ok := tb.(interface{ Parallel() });
		ok && spec.context.isParallel() {
		tb.Parallel()
	}

	t := newT(tb, spec.context)
	spec.printDescription(t)

	test := func(tb testing.TB) {
		defer spec.recoverFromPanic(tb)
		t := newT(tb, spec.context)
		defer t.setup()()
		blk(t)
	}

	if flakyFlag, isFlaky := spec.lookupFlaky(); isFlaky {
		at := Retry{Strategy: Waiter{WaitTimeout: flakyFlag.WaitTimeout}}
		at.Assert(tb, test)
	} else {
		test(tb)
	}
}

func (spec *Spec) recoverFromPanic(tb testing.TB) {
	if r := recover(); r != nil {
		_, file, line, _ := runtime.Caller(2)
		tb.Error(r, fmt.Sprintf(`%s:%d`, file, line), "\n", string(debug.Stack()))
	}
}

func (spec *Spec) runB(b *testing.B, blk func(*T)) {
	t := newT(b, spec.context)
	if _, ok := spec.lookupFlaky(); ok {
		b.Skip(`skipping flaky`)
	}

	for i := 0; i < b.N; i++ {
		func() {
			b.StopTimer()
			defer t.setup()()
			b.StartTimer()
			blk(t)
			b.StopTimer()
		}()
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type flakyFlag struct {
	WaitTimeout time.Duration
}
