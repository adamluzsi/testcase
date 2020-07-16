package testcase

import (
	"fmt"
	"reflect"
	"testing"
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
// 	go run ./... -vars -run "the/name/of/the/test/it/print/out/in/case/of/failure"
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
func (spec *Spec) Context(desc string, testContextBlock func(s *Spec)) {
	testContextBlock(spec.newSubSpec(desc))
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
func (spec *Spec) Test(desc string, test testCaseBlock) {
	spec.newSubSpec(desc).run(test)
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

const parallelWarn = `you can't use #Parallel after you already used when/and/then prior to calling Parallel`

// Parallel allows you to set all test case for the context where this is being called,
// and below to nested contexts, to be executed in parallel (concurrently).
// Keep in mind that you can call Parallel even from nested specs
// to apply Parallel testing for that context and below.
// This is useful when your test suite has no side effects at all.
// Using values from *vars when Parallel is safe.
// It is a shortcut for executing *testing.T#Parallel() for each test
func (spec *Spec) Parallel() {
	if spec.context.immutable {
		panic(parallelWarn)
	}

	spec.context.parallel = true
}

const sequentialWarn = `you can't use #Sequential after you already used when/and/then prior to calling Sequential`

// Sequential allows you to set all test case for the context where this is being called,
// and below to nested contexts, to be executed sequentially.
// It will negate any testcase.Spec#Parallel call effect.
// This is useful when you want to create a spec helper package
// and there you want to manage if you want to use components side effects or not.
func (spec *Spec) Sequential() {
	if spec.context.immutable {
		panic(sequentialWarn)
	}

	spec.context.sequential = true
}

const varWarning = `you cannot use let after a block is closed by a describe/when/and/then only before or within`

// Let define a memoized helper method.
// The value will be cached across multiple calls in the same example but not across examples.
// Note that Let is lazy-evaluated, it is not evaluated until the first time the method it defines is invoked.
// You can force this early by accessing the value from a Before block.
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
func (spec *Spec) Let(varName string, letBlock func(t *T) interface{}) {
	if spec.context.immutable {
		panic(varWarning)
	}

	spec.context.let(varName, letBlock)
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

const panicMessageForLetValue = `%T literal can't be used with #LetValue 
as the current implementation can't guarantee that the mutations on the value will not leak out to other tests,
please use the #Let memorization helper for now`

// LetValue is a shorthand for defining immutable vars with Let under the hood.
// So the function blocks can be skipped, which makes tests more readable.
func (spec *Spec) LetValue(varName string, value interface{}) {
	if _, ok := acceptedConstKind[reflect.ValueOf(value).Kind()]; !ok {
		panic(fmt.Sprintf(panicMessageForLetValue, value))
	}

	spec.Let(varName, func(t *T) interface{} {
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

func (spec *Spec) run(test func(t *T)) {
	if !spec.isAllowedToRun() {
		return
	}

	switch tb := spec.testingTB.(type) {
	case *testing.T:
		tb.Run(``, func(t *testing.T) {
			testCase := newT(t, spec.context)
			defer testCase.teardown()
			testCase.printDescription()
			testCase.setup()
			if spec.context.isParallel() {
				t.Parallel()
			}
			test(testCase)
		})
	case *testing.B:
		tb.Run(``, func(b *testing.B) {
			testCase := newT(b, spec.context)

			if testing.Verbose() {
				testCase.printDescription()
			}

			for i := 0; i < b.N; i++ {
				func() {
					b.StopTimer()
					testCase.reset()
					testCase.setup()
					defer testCase.teardown()

					b.StartTimer()
					test(testCase)
					b.StopTimer()
				}()
			}
		})
	default:
		panic(fmt.Errorf(`unsupported testing type: %T`, tb))
	}
}
