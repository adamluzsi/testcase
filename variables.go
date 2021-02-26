package testcase

import (
	"fmt"
	"strings"
)

func newVariables() *variables {
	return &variables{
		defs:  make(map[string]letBlock),
		cache: make(map[string]interface{}),
	}
}

// vars represents a set of vars for a given test spec
// Using the *vars object within the Then blocks/testCase edge cases is safe even when the *testing.T#parallel is called.
// One test case cannot leak its *vars object to another
type variables struct {
	defs  map[string]letBlock
	cache map[string]interface{}

	appliedOnLetHooks map[string]struct{}
}

func (v *variables) knows(varName string) bool {
	if _, found := v.defs[varName]; found {
		return true
	}
	if _, found := v.cache[varName]; found {
		return true
	}
	return false
}

func (v *variables) let(varName string, blk letBlock) {
	v.defs[varName] = blk
}

// I will return a testcase variable.
// it is suggested to use interface casting right after to it,
// so you can work with concrete types.
// If there is no such value, then it will panic with a "friendly" message.
func (v *variables) get(t *T, varName string) interface{} {
	t.TB.Helper()
	if !v.knows(varName) {
		panic(v.panicMessageFor(varName))
	}
	if _, found := v.cache[varName]; !found {
		v.cache[varName] = v.defs[varName](t)
	}
	return t.vars.cache[varName]
}

func (v *variables) set(varName string, value interface{}) {
	if _, ok := v.defs[varName]; !ok {
		v.let(varName, func(t *T) interface{} { return value })
	}
	v.cache[varName] = value
}

func (v *variables) reset() {
	v.cache = make(map[string]interface{})
}

func (v *variables) panicMessageFor(varName string) string {
	var messages []string
	messages = append(messages, fmt.Sprintf(`Variable %q is not found`, varName))
	var keys []string
	for k := range v.defs {
		keys = append(keys, k)
	}
	messages = append(messages, fmt.Sprintf(`Did you mean? %s`, strings.Join(keys, `, `)))
	return strings.Join(messages, ". ")
}

func (v *variables) merge(oth *variables) {
	for key, value := range oth.defs {
		v.defs[key] = value
	}
}

func (v *variables) addOnLetHookSetup(name string) {
	if v.appliedOnLetHooks == nil {
		v.appliedOnLetHooks = make(map[string]struct{})
	}
	v.appliedOnLetHooks[name] = struct{}{}
}

func (v *variables) hasOnLetHookApplied(name string) bool {
	if v.appliedOnLetHooks == nil {
		return false
	}

	_, ok := v.appliedOnLetHooks[name]
	return ok
}
