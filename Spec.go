package testcase

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func NewSpec(t *testing.T) *Spec {
	return newSubSpec(t, nil)
}

func newSubSpec(t *testing.T, parentCTX *context) *Spec {
	return &Spec{
		testingT:                    t,
		ctx:                         newTestCaseContext(parentCTX),
		preventContextConfiguration: false,
	}
}

// Spec provide you a struct that make building nested test context easy with the core T#Run function.
// ideal for synchronous nested test blocks
type Spec struct {
	testingT                    *testing.T
	ctx                         *context
	preventContextConfiguration bool
}

func (spec *Spec) Describe(subjectTopic string, specification func(s *Spec)) {
	spec.nest(`describe`, subjectTopic, specification)
}

func (spec *Spec) When(desc string, testContextBlock func(s *Spec)) {
	spec.nest(`when`, desc, testContextBlock)
}

func (spec *Spec) And(desc string, testContextBlock func(s *Spec)) {
	spec.nest(`when`, desc, testContextBlock)
}

func (spec *Spec) Then(desc string, test testCaseBlock) {
	spec.preventContextConfiguration = true

	spec.testingT.Run(desc, func(t *testing.T) {
		spec.runTestEdgeCase(t, test)
	})
}

// Before give you the ability to run a block before each test case.
// This is ideal for doing clean ahead before each test case.
// The received *testing.T object is the same as the Then block *testing.T object
// All setup block is stackable.
func (spec *Spec) Before(beforeBlock testCaseBlock) {
	spec.ctx.addHook(spec, func(t *testing.T, v *V) func() {
		beforeBlock(t, v)
		return func() {}
	})
}

// After give you the ability to run a block after each test case.
// This is ideal for running cleanups.
// The received *testing.T object is the same as the Then block *testing.T object
// All setup block is stackable.
func (spec *Spec) After(afterBlock testCaseBlock) {
	spec.ctx.addHook(spec, func(t *testing.T, v *V) func() {
		return func() { afterBlock(t, v) }
	})
}

// Around give you the ability to create "Before" setup for each test case,
// with the additional ability that the returned function will be defered to run after the Then block is done.
// This is ideal for setting up mocks, and then return the assertion request calls in the return func.
// All setup block is stackable.
func (spec *Spec) Around(aroundBlock hookBlock) {
	spec.ctx.addHook(spec, aroundBlock)
}

// Variables

func newV() *V {
	return &V{
		vars:  make(map[string]func(*V) interface{}),
		cache: make(map[string]interface{}),
	}
}

// V is represent a set of variable for a given test context
// the name is V only because it fits more nicely with the testing.T naming convention
// Using the *V object within the Then blocks / test edge cases is safe even when the *testing.T#Parallel is called.
// One test case cannot leak its *V object to another
type V struct {
	vars  map[string]func(*V) interface{}
	cache map[string]interface{}
}

const varWarning = `you cannot use let after a block is closed by a describe/when/and/then only before or within`

// Let allow you to define a test case variable.
// it is scoped to the current specification context/block,
// and cannot leak to higher level test cases.
//
// It is forbidden to use after a When/And/Then block,
// because then the current scope configuration is not homogen for all the edge cases.
// In order to prevent that, this will just simply panic with a warning message.
func (spec *Spec) Let(varName string, letBlock func(v *V) interface{}) {

	if spec.preventContextConfiguration {
		panic(varWarning)
	}

	spec.ctx.let(varName, letBlock)

}

// I will return a testcase variable.
// it is suggested to use interface casting right after to it,
// so you can work with concrete types.
// If there is no such value, then it will panic with a "friendly" message.
func (v *V) I(varName string) interface{} {
	fn, found := v.vars[varName]

	if !found {
		panic(v.panicMessageFor(varName))
	}

	if _, found := v.cache[varName]; !found {
		v.cache[varName] = fn(v)
	}

	return v.cache[varName]
}

// unexported

func (spec *Spec) runTestEdgeCase(t *testing.T, test func(t *testing.T, v *V)) {

	var teardown []func()

	v := newV()

	spec.ctx.eachLinkListElement(func(c *context) {
		v.merge(c.vars)
	})

	spec.ctx.eachLinkListElement(func(c *context) {
		for _, hook := range c.hooks {
			teardown = append(teardown, hook(t, v))
		}
	})

	defer func() {
		for _, td := range teardown {
			td()
		}
	}()

	test(t, v)

}

func (spec *Spec) nest(prefix, desc string, testContextBlock func(s *Spec)) {
	spec.preventContextConfiguration = true

	spec.testingT.Run(fmt.Sprintf(`%s %s`, prefix, desc), func(t *testing.T) {
		testContextBlock(newSubSpec(t, spec.ctx))
	})
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

type hookBlock func(*testing.T, *V) func()
type testCaseBlock func(*testing.T, *V)

func newTestCaseContext(parent *context) *context {
	return &context{
		hooks:  make([]hookBlock, 0),
		parent: parent,
		vars:   newV(),
	}
}

type contexts []*context

func (cs contexts) Len() int           { return len(cs) }
func (cs contexts) Less(i, j int) bool { return true }
func (cs contexts) Swap(i, j int)      { cs[i], cs[j] = cs[j], cs[i] }

type context struct {
	vars   *V
	parent *context
	hooks  []hookBlock
}

func (c *context) let(varName string, letBlock func(v *V) interface{}) {
	c.vars.vars[varName] = letBlock
}

func (c *context) eachLinkListElement(block func(*context)) {

	var (
		ctxs    contexts
		current *context
	)

	current = c

	for {
		ctxs = append(ctxs, current)

		if current.parent != nil {
			current = current.parent
			continue
		}

		break
	}

	sort.Sort(sort.Reverse(ctxs))

	for _, ctx := range ctxs {
		block(ctx)
	}

}

const hookWarning = `you cannot create spec hooks after you used describe/when/and/then,
unless you create a new context with the previously mentioned calls`

func (c *context) addHook(spec *Spec, h hookBlock) {
	if spec.preventContextConfiguration {
		panic(hookWarning)
	}

	c.hooks = append(c.hooks, h)
}
