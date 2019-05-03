package testrun

import (
	"fmt"
	"strings"
	"testing"
)

func NewSpec(t *testing.T) *Spec {
	return &Spec{testingT: t, preventContextConfiguration: false, stack: append(stack{}, newStage())}
}

// Spec provide you a struct that make building nested test context easy with the core T#Run function.
// ideal for synchronous nested test blocks
type Spec struct {
	testingT                    *testing.T
	stack                       stack
	preventContextConfiguration bool
}

func (spec *Spec) Describe(subjectTopic string, specification func(t *testing.T)) {
	spec.nest(`describe`, subjectTopic, specification)
}

func (spec *Spec) When(desc string, testContextBlock func(t *testing.T)) {
	spec.nest(`when`, desc, testContextBlock)
}

func (spec *Spec) And(desc string, testContextBlock func(t *testing.T)) {
	spec.nest(`when`, desc, testContextBlock)
}

func (spec *Spec) Then(desc string, test func(t *testing.T, v *V)) {
	spec.testingT.Run(desc, func(t *testing.T) {
		spec.runTestEdgeCase(t, test)
	})
	spec.preventContextConfiguration = true
}

// hooks

type Hook func(*testing.T) func()

func (spec *Spec) Before(beforeBlock func(t *testing.T)) {
	spec.stack.latestStage().addHook(spec, func(t *testing.T) func() {
		beforeBlock(t)
		return func() {}
	})
}

func (spec *Spec) After(afterBlock func(t *testing.T)) {
	spec.stack.latestStage().addHook(spec, func(t *testing.T) func() {
		return func() {
			afterBlock(t)
		}
	})
}

func (spec *Spec) Around(aroundBlock Hook) {
	spec.stack.latestStage().addHook(spec, aroundBlock)
}

// Variables

func newV() *V {
	return &V{vars: make(vars)}
}

// V is represent a set of variable for a given test context
// the name is V only because it fits more nicely with the testing.T naming convention
type V struct{ vars }

type vars map[string]func(*V) interface{}

const varWarning = `you cannot use let after a block is closed by a describe/when/and/then only before or within`

func (spec *Spec) Let(varName string, letBlock func(v *V) interface{}) {

	if spec.preventContextConfiguration {
		panic(varWarning)
	}

	spec.stack.latestStage().v.vars[varName] = letBlock
}

func (v *V) I(varName string) interface{} {
	fn, found := v.vars[varName]

	if !found {
		panic(v.panicMessageFor(varName))
	}

	return fn(v)
}

// unexported

func (spec *Spec) runTestEdgeCase(t *testing.T, test func(t *testing.T, v *V)) {

	var teardown []func()

	v := newV()

	for _, stage := range spec.stack {
		for _, hook := range stage.hooks {
			teardown = append(teardown, hook(t))
		}

		for k, f := range stage.v.vars {
			v.vars[k] = f
		}
	}

	defer func() {
		for _, td := range teardown {
			td()
		}
	}()

	test(t, v)

}

func (spec *Spec) nest(prefix, desc string, testContextBlock func(t *testing.T)) {
	defer spec.nextStage()()
	spec.testingT.Run(fmt.Sprintf(`%s %s`, prefix, desc), testContextBlock)
}

func (spec *Spec) nextStage() func() {
	spec.stack = append(spec.stack, newStage())
	spec.preventContextConfiguration = false

	return func() {
		spec.stack = spec.stack[:len(spec.stack)-1]
		spec.preventContextConfiguration = true
	}
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

type stack []*stage

func (s stack) latestStage() *stage {
	return s[len(s)-1]
}

func newStage() *stage {
	return &stage{
		hooks: make([]Hook, 0),
		v:     newV(),
	}
}

type stage struct {
	hooks []Hook
	v     *V
}

const hookWarning = `you cannot create spec hooks after you used describe/when/and/then,
unless you create a new context with the previously mentioned calls`

func (s *stage) addHook(spec *Spec, h Hook) {

	if spec.preventContextConfiguration {
		panic(hookWarning)
	}

	s.hooks = append(s.hooks, h)
}
