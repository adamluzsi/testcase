package testcase

import (
	"fmt"
	"strings"
	"sync"
)

func newVariables() *variables {
	return &variables{
		defs:  make(map[string]letBlock),
		cache: make(map[string]interface{}),
		onLet: make(map[string]struct{}),
		locks: make(map[string]*sync.RWMutex),
	}
}

// variables represents an individual test case's runtime variables.
// Using the variables cache within the individual test cases are safe even with *testing#T.Parallel().
// Different test cases don't share they variables instance.
type variables struct {
	m     sync.RWMutex
	defs  map[string]letBlock
	cache map[string]interface{}
	onLet map[string]struct{}
	locks map[string]*sync.RWMutex
}

func (v *variables) Knows(varName string) bool {
	defer v.rLock(varName)()
	if _, found := v.defs[varName]; found {
		return true
	}
	if _, found := v.cache[varName]; found {
		return true
	}
	return false
}

func (v *variables) Let(varName string, blk letBlock /* [interface{}] */) {
	defer v.lock(varName)()
	v.let(varName, blk)
}

func (v *variables) let(varName string, blk letBlock /* [interface{}] */) {
	v.defs[varName] = blk
}

// Get will return a testcase vs.
//
// If there is no such value, then it will panic with a "friendly" message.
func (v *variables) Get(t *T, varName string) interface{} {
	t.TB.Helper()
	if !v.Knows(varName) {
		t.Fatal(v.fatalMessageFor(varName))
	}
	defer v.lock(varName)()
	if _, found := v.cache[varName]; !found {
		v.cache[varName] = v.defs[varName](t)
	}
	return t.vars.cache[varName]
}

func (v *variables) Set(varName string, value interface{}) {
	defer v.lock(varName)()
	if _, ok := v.defs[varName]; !ok {
		v.let(varName, func(t *T) interface{} { return value })
	}
	v.cache[varName] = value
}

func (v *variables) reset() {
	v.cache = make(map[string]interface{})
}

func (v *variables) fatalMessageFor(varName string) string {
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
	v.onLet[name] = struct{}{}
}

func (v *variables) hasOnLetHookApplied(name string) bool {
	_, ok := v.onLet[name]
	return ok
}

func (v *variables) rLock(varName string) func() {
	m := v.mutex(varName)
	m.Lock()
	return m.Unlock
}

func (v *variables) lock(varName string) func() {
	m := v.mutex(varName)
	m.Lock()
	return m.Unlock
}

func (v *variables) mutex(varName string) *sync.RWMutex {
	v.m.Lock()
	defer v.m.Unlock()
	if _, ok := v.locks[varName]; !ok {
		v.locks[varName] = &sync.RWMutex{}
	}
	return v.locks[varName]
}
