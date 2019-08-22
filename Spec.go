package testcase

import (
	"fmt"
	"strings"
	"testing"
)

// NewSpec create new Spec struct that is ready for usage.
func NewSpec(t *testing.T) *Spec {
	return &Spec{
		testingT: t,
		ctx:      newContext(),
	}
}

// T embeds both testcase variables, and testing#T functionality.
// This leave place open for extension and
// but define a stable foundation for the hooks and test edge case function signatures
//
// Works as a drop in replacement for pkgs where they depend on one of the function of testing#T
//
type T struct {
	*testing.T
	*V
}

// I will return a testcase variable.
// it is suggested to use interface casting right after to it,
// so you can work with concrete types.
// If there is no such value, then it will panic with a "friendly" message.
func (t *T) I(varName string) interface{} {
	fn, found := t.V.vars[varName]

	if !found {
		panic(t.V.panicMessageFor(varName))
	}

	if _, found := t.V.cache[varName]; !found {
		t.V.cache[varName] = fn(t)
	}

	return t.V.cache[varName]
}

// Spec provides you a struct that makes building nested test context easy with the core T#Context function.
//
// spec structure is a simple wrapping around the testing.T#Context.
// It doesn't use any global singleton cache object or anything like that.
// It doesn't force you to use global variables.
//
// It uses the same idiom as the core go testing pkg also provide you.
// You can use the same way as the core testing pkg
// 	go run ./... -v -run "the/name/of/the/test/it/print/out/in/case/of/failure"
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
// Using values from *V when Parallel is safe.
// It is a shortcut for executing *testing.T#Parallel() for each test
func (spec *Spec) Parallel() {

	if spec.ctx.immutable {
		panic(parallelWarn)
	}

	spec.ctx.parallel = true
}

const varWarning = `you cannot use let after a block is closed by a describe/when/and/then only before or within`

// Let allow you to define a test case variable to a given scope, and below scopes.
// It cannot leak to higher level scopes, and between Concurrent test runs.
// Calling Let in a nested/sub scope will apply the new value for that value to that scope and below.
//
// It will panic if it is used after a When/And/Then scope definition,
// because those scopes would have no clue about the later defined variable.
// In order to keep the specification reading mental model requirement low,
// it is intentionally not implemented to handle such case.
//
// variables strictly belong to a given `Describe`/`When`/`And` scope,
// and configured before any hook would be applied,
// therefore hooks always receive the most latest version from the `Let` variables,
// regardless in which scope the hook that use the varable is define.
//
func (spec *Spec) Let(varName string, letBlock func(t *T) interface{}) {
	if spec.ctx.immutable {
		panic(varWarning)
	}

	spec.ctx.let(varName, letBlock)
}

func (spec *Spec) runTestCase(test func(t *T)) {

	allCTX := spec.ctx.allLinkListElement()

	var desc []string

	for _, c := range allCTX[1:] {
		desc = append(desc, c.description)
	}

	spec.testingT.Run(strings.Join(desc, `_`), func(runT *testing.T) {

		v := newV()
		t := &T{T: runT, V: v}
		var teardown []func()

		spec.printDescription(t)

		for _, c := range allCTX {
			v.merge(c.vars)
		}

		for _, c := range allCTX {
			for _, hook := range c.hooks {
				teardown = append(teardown, hook(t))
			}
		}

		defer func() {
			for _, td := range teardown {
				td()
			}
		}()

		if spec.ctx.isParallel() {
			t.Parallel()
		}

		test(t)

	})
}

func newV() *V {
	return &V{
		vars:  make(map[string]func(*T) interface{}),
		cache: make(map[string]interface{}),
	}
}

// V represents a set of variables for a given test context
// the name is V only because it fits more nicely with the testing.T naming convention
// Using the *V object within the Then blocks/test edge cases is safe even when the *testing.T#Parallel is called.
// One test case cannot leak its *V object to another
type V struct {
	vars  map[string]func(*T) interface{}
	cache map[string]interface{}
}

func (v *V) panicMessageFor(varName string) string {

	var msgs []string
	msgs = append(msgs, fmt.Sprintf(`Variable %q is not found`, varName))

	var keys []string
	for k := range v.vars {
		keys = append(keys, k)
	}

	msgs = append(msgs, fmt.Sprintf(`Did you mean? %s`, strings.Join(keys, `, `)))

	return strings.Join(msgs, ". ")

}

func (v *V) merge(oth *V) {
	for key, value := range oth.vars {
		v.vars[key] = value
	}
}

type hookBlock func(*T) func()
type testCaseBlock func(*T)

func newContext() *context {
	return &context{
		hooks:     make([]hookBlock, 0),
		parent:    nil,
		vars:      newV(),
		immutable: false,
	}
}

type context struct {
	vars        *V
	parent      *context
	hooks       []hookBlock
	parallel    bool
	immutable   bool
	description string
}

func (c *context) let(varName string, letBlock func(*T) interface{}) {
	c.vars.vars[varName] = letBlock
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

	log(t, lines...)
}
