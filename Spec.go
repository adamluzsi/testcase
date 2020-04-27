package testcase

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func newT(runT *testing.T) *T {
	return &T{T: runT, variables: newVariables()}
}

// T embeds both testcase variables, and testing#T functionality.
// This leave place open for extension and
// but define a stable foundation for the hooks and test edge case function signatures
//
// Works as a drop in replacement for packages where they depend on one of the function of testing#T
//
type T struct {
	*testing.T
	variables *variables
	defers    []func()
}

// I will return a testcase variable.
// it is suggested to use interface casting right after to it,
// so you can work with concrete types.
// If there is no such value, then it will panic with a "friendly" message.
func (t *T) I(varName string) interface{} {
	return t.variables.get(t, varName)
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
	t.variables.set(varName, value)
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
// In a practical example, this means that if you have common variables defined with testcase.Spec#Let memorization,
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
		if expected := rfnType.In(i).Kind(); expected != value.Kind() {
			_, file, line, _ := runtime.Caller(1)
			const format = "deferred function argument[%d] type mismatch: expected %s, but got %s from %s:%d"
			panic(fmt.Sprintf(format, i, expected, value.Kind(), file, line))
		}
		rargs = append(rargs, value)
	}
	t.defers = append(t.defers, func() { rfn.Call(rargs) })
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

// NewSpec create new Spec struct that is ready for usage.
func NewSpec(t *testing.T) *Spec {
	return &Spec{
		testingT: t,
		ctx:      newContext(),
	}
}

// Spec provides you a struct that makes building nested test context easy with the core T#Context function.
//
// spec structure is a simple wrapping around the testing.T#Context.
// It doesn't use any global singleton cache object or anything like that.
// It doesn't force you to use global variables.
//
// It uses the same idiom as the core go testing pkg also provide you.
// You can use the same way as the core testing pkg
// 	go run ./... -variables -run "the/name/of/the/test/it/print/out/in/case/of/failure"
//
// It allows you to do context preparation for each test in a way,
// that it will be safe for use with testing.T#Parallel.
type Spec struct {
	testingT *testing.T
	ctx      *context
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

// Test creates a test case block where you receive the fully configured `testcase#T` object.
// Hook contents that meant to run before the test edge cases will run before the function the Test receives,
// and hook contents that meant to run after the test edge cases will run after the function is done.
// After hooks are deferred after the received function block, so even in case of panic, it will still be executed.
//
// It should not contain anything that modify the test subject input.
// It should focuses only on asserting the result of the subject.
//
func (spec *Spec) Test(desc string, test testCaseBlock) {
	spec.newSubSpec(desc).runTestCase(test)
}

// Before give you the ability to run a block before each test case.
// This is ideal for doing clean ahead before each test case.
// The received *testing.T object is the same as the Test block *testing.T object
// This hook applied to this scope and anything that is nested from here.
// All setup block is stackable.
func (spec *Spec) Before(beforeBlock testCaseBlock) {
	spec.ctx.addHook(func(t *T) func() {
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
	spec.ctx.addHook(func(t *T) func() {
		return func() { afterBlock(t) }
	})
}

// Around give you the ability to create "Before" setup for each test case,
// with the additional ability that the returned function will be deferred to run after the Then block is done.
// This is ideal for setting up mocks, and then return the assertion request calls in the return func.
// This hook applied to this scope and anything that is nested from here.
// All setup block is stackable.
func (spec *Spec) Around(aroundBlock hookBlock) {
	spec.ctx.addHook(aroundBlock)
}

const parallelWarn = `you cannot use #Parallel after you already used when/and/then prior to calling Parallel`

// Parallel allows you to set all test case for the context where this is being called,
// and below to nested contexts, to be executed in parallel (concurrently).
// Keep in mind that you can call Parallel even from nested specs
// to apply Parallel testing for that context and below.
// This is useful when your test suite has no side effects at all.
// Using values from *variables when Parallel is safe.
// It is a shortcut for executing *testing.T#Parallel() for each test
func (spec *Spec) Parallel() {
	if spec.ctx.immutable {
		panic(parallelWarn)
	}

	spec.ctx.parallel = true
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
// Defining test variables always expected in the beginning of a specification scope,
// mainly for readability reasons.
//
// variables strictly belong to a given `Describe`/`When`/`And` scope,
// and configured before any hook would be applied,
// therefore hooks always receive the most latest version from the `Let` variables,
// regardless in which scope the hook that use the variable is define.
//
func (spec *Spec) Let(varName string, letBlock func(t *T) interface{}) {
	if spec.ctx.immutable {
		panic(varWarning)
	}

	spec.ctx.let(varName, letBlock)
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

// LetValue is a shorthand for defining immutable variables with Let under the hood.
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

func (spec *Spec) runTestCase(test func(t *T)) {

	allCTX := spec.ctx.allLinkListElement()

	var desc []string

	for _, c := range allCTX[1:] {
		desc = append(desc, c.description)
	}

	spec.testingT.Run(strings.Join(desc, `_`), func(runT *testing.T) {
		t := newT(runT)
		defer t.teardown()

		spec.printDescription(t)

		for _, c := range allCTX {
			t.variables.merge(c.vars)
		}

		for _, c := range allCTX {
			for _, hook := range c.hooks {
				t.Defer(hook(t))
			}
		}

		if spec.ctx.isParallel() {
			t.Parallel()
		}

		test(t)

	})
}

func newVariables() *variables {
	return &variables{
		defs:  make(map[string]func(*T) interface{}),
		cache: make(map[string]interface{}),
	}
}

// variables represents a set of variables for a given test context
// Using the *variables object within the Then blocks/test edge cases is safe even when the *testing.T#Parallel is called.
// One test case cannot leak its *variables object to another
type variables struct {
	defs  map[string]func(*T) interface{}
	cache map[string]interface{}
}

// I will return a testcase variable.
// it is suggested to use interface casting right after to it,
// so you can work with concrete types.
// If there is no such value, then it will panic with a "friendly" message.
func (v *variables) get(t *T, varName string) interface{} {
	fn, found := v.defs[varName]

	if !found {
		panic(v.panicMessageFor(varName))
	}

	if _, found := v.cache[varName]; !found {
		v.cache[varName] = fn(t)
	}

	return t.variables.cache[varName]
}

func (v *variables) set(varName string, value interface{}) {
	if _, ok := v.defs[varName]; !ok {
		v.defs[varName] = func(t *T) interface{} { return value }
	}
	v.cache[varName] = value
}

func (v *variables) panicMessageFor(varName string) string {

	var msgs []string
	msgs = append(msgs, fmt.Sprintf(`Variable %q is not found`, varName))

	var keys []string
	for k := range v.defs {
		keys = append(keys, k)
	}

	msgs = append(msgs, fmt.Sprintf(`Did you mean? %s`, strings.Join(keys, `, `)))

	return strings.Join(msgs, ". ")

}

func (v *variables) merge(oth *variables) {
	for key, value := range oth.defs {
		v.defs[key] = value
	}
}

type hookBlock func(*T) func()
type testCaseBlock func(*T)

func newContext() *context {
	return &context{
		hooks:     make([]hookBlock, 0),
		parent:    nil,
		vars:      newVariables(),
		immutable: false,
	}
}

type context struct {
	vars        *variables
	parent      *context
	hooks       []hookBlock
	parallel    bool
	immutable   bool
	description string
}

func (c *context) let(varName string, letBlock func(*T) interface{}) {
	c.vars.defs[varName] = letBlock
}

func (c *context) isParallel() bool {
	for _, ctx := range c.allLinkListElement() {
		if ctx.parallel {
			return true
		}
	}
	return false
}

func (c *context) allLinkListElement() []*context {
	var (
		contexts []*context
		current  *context
	)

	current = c

	for {

		contexts = append([]*context{current}, contexts...)

		if current.parent != nil {
			current = current.parent
			continue
		}

		break
	}

	return contexts
}

const hookWarning = `you cannot create spec hooks after you used describe/when/and/then,
unless you create a new context with the previously mentioned calls`

func (c *context) addHook(h hookBlock) {
	if c.immutable {
		panic(hookWarning)
	}

	c.hooks = append(c.hooks, h)
}

func (spec *Spec) newSubSpec(desc string) *Spec {
	spec.ctx.immutable = true
	subCTX := newContext()
	subCTX.parent = spec.ctx
	subCTX.description = desc
	subSpec := &Spec{testingT: spec.testingT, ctx: subCTX}
	return subSpec
}

func (spec *Spec) printDescription(t *T) {
	var lines []interface{}

	var spaceIndentLevel int
	for _, c := range spec.ctx.allLinkListElement() {
		if c.description == `` {
			continue
		}

		lines = append(lines, fmt.Sprintln(strings.Repeat(` `, spaceIndentLevel*2), c.description))
		spaceIndentLevel++
	}

	log(t.T, lines...)
}
