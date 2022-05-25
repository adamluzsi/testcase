package testcase

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal"
)

// NewSpec create new Spec struct that is ready for usage.
func NewSpec(tb testing.TB, opts ...SpecOption) *Spec {
	tb.Helper()
	var s *Spec
	switch tb := tb.(type) {
	case *T:
		s = tb.spec.newSubSpec("", opts...)
	default:
		s = newSpec(tb, opts...)
		s.seed = seedForSpec(tb)
		s.orderer = newOrderer(tb, s.seed)
		tb.Cleanup(s.Finish)
	}
	return s
}

func newSpec(tb testing.TB, opts ...SpecOption) *Spec {
	tb.Helper()
	s := &Spec{
		testingTB: tb,
		vars:      newVariables(),
		immutable: false,
	}
	for _, to := range opts {
		to.setup(s)
	}
	return s
}

func (spec *Spec) newSubSpec(desc string, opts ...SpecOption) *Spec {
	spec.testingTB.Helper()
	spec.immutable = true
	sub := newSpec(spec.testingTB, opts...)
	sub.parent = spec
	sub.seed = spec.seed
	sub.orderer = spec.orderer
	sub.description = desc
	spec.children = append(spec.children, sub)
	return sub
}

// Spec provides you a struct that makes building nested test spec easy with the core T#Context function.
//
// spec structure is a simple wrapping around the testing.T#Context.
// It doesn't use any global singleton cache object or anything like that.
// It doesn't force you to use global vars.
//
// It uses the same idiom as the core go testing pkg also provide you.
// You can use the same way as the core testing pkg
// 	go run ./... -v -run "the/name/of/the/test/it/print/orderingOutput/in/case/of/failure"
//
// It allows you to do spec preparation for each test in a way,
// that it will be safe for use with testing.T#Parallel.
type Spec struct {
	testingTB testing.TB

	parent   *Spec
	children []*Spec

	hooks struct {
		Around    []hookBlock
		AroundAll []func() func()
	}

	immutable     bool
	vars          *variables
	parallel      bool
	sequential    bool
	skipBenchmark bool
	flaky         *Eventually
	eventually    *Eventually
	group         *struct{ name string }
	description   string
	tags          []string
	tests         []func()
	finished      bool
	orderer       orderer
	seed          int64
}

// Context allow you to create a sub specification for a given spec.
// In the sub-specification it is expected to add more contextual information to the test
// in a form of hook of variable setting.
// With Context you can set your custom test description, without any forced prefix like describe/when/and.
//
// It is basically piggybacking the testing#T.Context and create new subspec in that nested testing#T.Context scope.
// It is used to add more description spec for the given subject.
// It is highly advised to always use When + Before/Around together,
// in which you should setup exactly what you wrote in the When description input.
// You can Context as many When/And within each other, as you want to achieve
// the most concrete edge case you want to test.
//
// To verify easily your state-machine, you can count the `if`s in your implementation,
// and check that each `if` has 2 `When` block to represent the two possible path.
//
func (spec *Spec) Context(desc string, testContextBlock contextBlock, opts ...SpecOption) {
	spec.testingTB.Helper()
	sub := spec.newSubSpec(desc, opts...)

	// when no new group defined
	if sub.group == nil {
		testContextBlock(sub)
		return
	}

	name := escapeName(sub.group.name)
	switch tb := spec.testingTB.(type) {
	case tRunner:
		tb.Run(name, func(t *testing.T) {
			sub.withFinishUsingTestingTB(t, func() {
				testContextBlock(sub)
			})
		})

	case bRunner:
		tb.Run(name, func(b *testing.B) {
			sub.withFinishUsingTestingTB(b, func() {
				testContextBlock(sub)
			})
		})

	case TBRunner:
		tb.Run(name, func(tb testing.TB) {
			sub.withFinishUsingTestingTB(tb, func() {
				testContextBlock(sub)
			})
		})

	default:
		testContextBlock(sub)
	}
}

type contextBlock func(s *Spec)

type block func(*T)

// Test creates a test case block where you receive the fully configured `testcase#T` object.
// Hook contents that meant to run before the test edge cases will run before the function the Test receives,
// and hook contents that meant to run after the test edge cases will run after the function is done.
// After hooks are deferred after the received function block, so even in case of panic, it will still be executed.
//
// It should not contain anything that modify the test subject input.
// It should focuses only on asserting the result of the subject.
//
func (spec *Spec) Test(desc string, test block, opts ...SpecOption) {
	spec.testingTB.Helper()
	s := spec.newSubSpec(desc, opts...)
	s.run(test)
}

const warnEventOnImmutableFormat = `you can't use .%s after you already used when/and/then`

// Parallel allows you to set list test case for the spec where this is being called,
// and below to nested contexts, to be executed in parallel (concurrently).
// Keep in mind that you can call Parallel even from nested specs
// to apply Parallel testing for that spec and below.
// This is useful when your test suite has no side effects at list.
// Using values from *vars when Parallel is safe.
// It is a shortcut for executing *testing.T#Parallel() for each test
func (spec *Spec) Parallel() {
	spec.testingTB.Helper()
	if spec.immutable {
		spec.testingTB.Fatalf(warnEventOnImmutableFormat, `Parallel`)
	}
	parallel().setup(spec)
}

// SkipBenchmark will flag the current Spec / Context to be skipped during Benchmark mode execution.
// If you wish to skip only a certain test, not the whole Spec / Context, use the SkipBenchmark SpecOption instead.
func (spec *Spec) SkipBenchmark() {
	spec.testingTB.Helper()
	if spec.immutable {
		spec.testingTB.Fatalf(warnEventOnImmutableFormat, `SkipBenchmark`)
	}
	SkipBenchmark().setup(spec)
}

// Sequential allows you to set list test case for the spec where this is being called,
// and below to nested contexts, to be executed sequentially.
// It will negate any testcase.Spec#Parallel call effect.
// This is useful when you want to create a spec helper package
// and there you want to manage if you want to use components side effects or not.
func (spec *Spec) Sequential() {
	spec.testingTB.Helper()
	if spec.immutable {
		panic(fmt.Sprintf(warnEventOnImmutableFormat, `Sequential`))
	}
	sequential().setup(spec)
}

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
	spec.testingTB.Helper()
	spec.tags = append(spec.tags, tags...)
}

func (spec *Spec) isAllowedToRun() bool {
	spec.testingTB.Helper()
	currentTagSet := spec.getTagSet()
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
	spec.testingTB.Helper()
	for _, context := range spec.list() {
		if context.skipBenchmark {
			return false
		}
	}
	return true
}

func (spec *Spec) lookupRetryFlaky() (Eventually, bool) {
	spec.testingTB.Helper()
	for _, context := range spec.list() {
		if context.flaky != nil {
			return *context.flaky, true
		}
	}
	return Eventually{}, false
}

func (spec *Spec) lookupRetryEventually() (Eventually, bool) {
	spec.testingTB.Helper()
	for _, context := range spec.list() {
		if context.eventually != nil {
			return *context.eventually, true
		}
	}
	return Eventually{}, false
}

func (spec *Spec) printDescription(t *T) {
	spec.testingTB.Helper()
	var lines []interface{}

	var spaceIndentLevel int
	for _, c := range t.contexts() {
		if c.description == `` {
			continue
		}

		lines = append(lines, fmt.Sprintln(strings.Repeat(` `, spaceIndentLevel*2), c.description))
		spaceIndentLevel++
	}

	internal.Log(t, lines...)
}

// TODO: add group name representation here
func (spec *Spec) name() string {
	var desc string
	for _, context := range spec.list() {
		if desc != `` {
			desc += ` `
		}

		if context.group == nil {
			desc += context.description
		}
	}
	name := escapeName(desc)
	if name == `` {
		name = internal.CallerLocation(3, true)
	}
	return name
}

///////////////////////////////////////////////////////=- run -=////////////////////////////////////////////////////////

func (spec *Spec) run(blk func(*T)) {
	spec.testingTB.Helper()
	if !spec.isAllowedToRun() {
		return
	}
	name := spec.name()
	switch tb := spec.testingTB.(type) {
	case tRunner:
		spec.addTest(func() {
			tb.Run(name, func(t *testing.T) {
				t.Helper()
				spec.runTB(t, blk)
			})
		})
	case bRunner:
		if !spec.isBenchAllowedToRun() {
			return
		}
		spec.addTest(func() {
			tb.Run(name, func(b *testing.B) {
				b.Helper()
				spec.runB(b, blk)
			})
		})
	case TBRunner:
		spec.addTest(func() {
			tb.Run(name, func(tb testing.TB) {
				tb.Helper()
				spec.runTB(tb, blk)
			})
		})
	default:
		spec.addTest(func() {
			tb.Helper()
			spec.runTB(tb, blk)
		})
	}
}

func (spec *Spec) runTB(tb testing.TB, blk func(*T)) {
	spec.testingTB.Helper()
	tb.Helper()
	if tb, ok := tb.(interface{ Parallel() }); ok && spec.isParallel() {
		tb.Parallel()
	}

	spec.printDescription(newT(tb, spec))

	test := func(tb testing.TB) {
		tb.Helper()
		defer spec.recoverFromPanic(tb)
		t := newT(tb, spec)
		defer t.setUp()()
		blk(t)
	}

	retryHandler, ok := spec.lookupRetryFlaky()
	if ok {
		retryHandler.Assert(tb, func(it assert.It) { test(it) })
	} else {
		test(tb)
	}
}

func (spec *Spec) recoverFromPanic(tb testing.TB) {
	spec.testingTB.Helper()
	tb.Helper()
	if r := recover(); r != nil {
		_, file, line, _ := runtime.Caller(2)
		tb.Error(r, fmt.Sprintf(`%s:%d`, file, line), "\n", string(debug.Stack()))
	}
}

func (spec *Spec) runB(b *testing.B, blk func(*T)) {
	spec.testingTB.Helper()
	b.Helper()
	t := newT(b, spec)
	if _, ok := spec.lookupRetryFlaky(); ok {
		b.Skip(`skipping because flaky flag`)
	}
	benchCase := func() {
		b.StopTimer()
		b.Helper()
		defer t.setUp()()
		b.StartTimer()
		defer b.StopTimer()
		blk(t)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchCase()
	}
}

func (spec *Spec) acceptVisitor(v visitor) {
	for _, child := range spec.children {
		child.acceptVisitor(v)
	}
	v.Visit(spec)
}

// Finish executes all unfinished test and mark them finished.
// Finish can be used when it is important to run the test before the Spec's testing#TB.Cleanup would execute.
//
// Such case can be when a resource leaked inside a testing scope
// and resource closed with a deferred function, but the spec is still not ran.
func (spec *Spec) Finish() {
	spec.testingTB.Helper()
	var tests []func()
	var hooks []func() func()
	spec.acceptVisitor(visitorFunc(func(s *Spec) {
		if s.finished {
			return
		}
		s.finished = true
		s.immutable = true
		tests = append(tests, s.tests...)
		hooks = append(hooks, s.hooks.AroundAll...)
	}))
	spec.orderer.Order(tests)
	td := &internal.Teardown{}
	defer td.Finish()
	for _, hook := range hooks {
		td.Defer(hook())
	}
	for _, tc := range tests {
		tc()
	}
}

func (spec *Spec) withFinishUsingTestingTB(tb testing.TB, blk func()) {
	spec.testingTB.Helper()
	tb.Helper()
	ogTB := spec.testingTB
	defer func() { spec.testingTB = ogTB }()
	spec.testingTB = tb
	blk()
	spec.Finish()
}

func (spec *Spec) isParallel() bool {
	spec.testingTB.Helper()
	var (
		isParallel   bool
		isSequential bool
	)

	for _, ctx := range spec.list() {
		if ctx.parallel {
			isParallel = true
		}
		if ctx.sequential {
			isSequential = true
		}
	}

	return isParallel && !isSequential
}

// visits *Spec chain in a reverse order
// from parent till the current children.
func (spec *Spec) list() []*Spec {
	var (
		specs   []*Spec
		current = spec
	)
	for {
		specs = append([]*Spec{current}, specs...)
		if current.parent != nil {
			current = current.parent
			continue
		}
		break
	}
	return specs
}

func (spec *Spec) getTagSet() map[string]struct{} {
	spec.testingTB.Helper()
	tagsSet := make(map[string]struct{})
	for _, ctx := range spec.list() {
		for _, tag := range ctx.tags {
			tagsSet[tag] = struct{}{}
		}
	}
	return tagsSet
}

func (spec *Spec) addTest(blk func()) {
	spec.testingTB.Helper()
	spec.tests = append(spec.tests, blk)
}

var escapeNameRGX = regexp.MustCompile(`\\.`)

func escapeName(s string) string {
	const charsToEscape = `.,'";`
	for _, char := range charsToEscape {
		s = strings.Replace(s, string(char), ``, -1)
	}
	s = regexp.QuoteMeta(s)
	for _, esc := range escapeNameRGX.FindAllStringSubmatch(s, -1) {
		s = strings.Replace(s, esc[0], ``, -1)
	}
	return s
}
