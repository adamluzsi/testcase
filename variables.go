package testcase

import (
	"fmt"
	"strings"
)

func newVariables() *variables {
	return &variables{
		defs:  make(map[string]func(*T) interface{}),
		cache: make(map[string]interface{}),
	}
}

// vars represents a set of vars for a given test context
// Using the *vars object within the Then blocks/test edge cases is safe even when the *testing.T#Parallel is called.
// One test case cannot leak its *vars object to another
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

	return t.vars.cache[varName]
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
