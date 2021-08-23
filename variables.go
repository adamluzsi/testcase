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
	mutex sync.RWMutex
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
	if !v.cacheHas(varName) {
		// cacheSet(varName, ...) is protected from concurrent access by lock(varName).
		v.cacheSet(varName, v.defs[varName](t))
	}
	return t.vars.cacheGet(varName)
}

func (v *variables) cacheGet(varName string) interface{} {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.cache[varName]
}

func (v *variables) cacheHas(varName string) bool {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	_, ok := v.cache[varName]
	return ok
}

func (v *variables) cacheSet(varName string, data interface{}) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.cache[varName] = data
}

func (v *variables) Set(varName string, value interface{}) {
	defer v.lock(varName)()
	if _, ok := v.defs[varName]; !ok {
		v.let(varName, func(t *T) interface{} { return value })
	}
	v.cacheSet(varName, value)
}

func (v *variables) reset() {
	v.mutex.Lock()
	defer v.mutex.Unlock()
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
	m := v.getMutex(varName)
	m.RLock()
	return m.RUnlock
}

func (v *variables) lock(varName string) func() {
	m := v.getMutex(varName)
	m.Lock()
	return m.Unlock
}

func (v *variables) getMutex(varName string) *sync.RWMutex {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if _, ok := v.locks[varName]; !ok {
		v.locks[varName] = &sync.RWMutex{}
	}
	return v.locks[varName]
}
