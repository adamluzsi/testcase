package testcase

import (
	"fmt"
	"strings"
	"sync"
)

func newVariables() *variables {
	return &variables{
		defs:       make(map[VarID]variablesInitBlock),
		defsSuper:  make(map[VarID][]variablesInitBlock),
		cache:      make(map[VarID]interface{}),
		cacheSuper: newVariablesSuperCache(),
		onLet:      make(map[VarID]struct{}),
		locks:      make(map[VarID]*sync.RWMutex),
		before:     make(map[VarID]struct{}),
		deps:       make(map[VarID]*sync.Once),
	}
}

// variables represents an individual test case's runtime variables.
// Using the variables cache within the individual test cases are safe even with *testing#T.Parallel().
// Different test cases don't share they variables instance.
type variables struct {
	mutex      sync.RWMutex
	locks      map[VarID]*sync.RWMutex
	defs       map[VarID]variablesInitBlock
	defsSuper  map[VarID][]variablesInitBlock
	onLet      map[VarID]struct{}
	before     map[VarID]struct{}
	cache      map[VarID]any
	cacheSuper *variablesSuperCache
	deps       map[VarID]*sync.Once
}

type variablesInitBlock func(t *T) any

func (v *variables) Knows(varName VarID) bool {
	defer v.rLock(varName)()
	if _, found := v.defs[varName]; found {
		return true
	}
	if _, found := v.cache[varName]; found {
		return true
	}
	return false
}

func (v *variables) Let(varName VarID, blk variablesInitBlock /* [interface{}] */) {
	defer v.lock(varName)()
	v.let(varName, blk)
}

func (v *variables) let(varName VarID, blk variablesInitBlock /* [interface{}] */) {
	v.defs[varName] = blk
}

// Get will return a testcase vs.
//
// If there is no such value, then it will panic with a "friendly" message.
func (v *variables) Get(t *T, varName VarID) interface{} {
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

func (v *variables) cacheGet(varName VarID) interface{} {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.cache[varName]
}

func (v *variables) cacheHas(varName VarID) bool {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	_, ok := v.cache[varName]
	return ok
}

func (v *variables) cacheSet(varName VarID, data interface{}) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.cache[varName] = data
}

func (v *variables) Set(varName VarID, value interface{}) {
	defer v.lock(varName)()
	if _, ok := v.defs[varName]; !ok {
		v.let(varName, func(t *T) interface{} { return value })
	}
	v.cacheSet(varName, value)
}

func (v *variables) reset() {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.cache = make(map[VarID]interface{})
	v.cacheSuper = newVariablesSuperCache()
}

func (v *variables) fatalMessageFor(varName VarID) string {
	var messages []string
	messages = append(messages, fmt.Sprintf(`Variable %q is not found`, varName))
	var keys []VarID
	for k := range v.defs {
		keys = append(keys, k)
	}
	messages = append(messages, `Did you mean?`)
	for _, vn := range keys {
		messages = append(messages, fmt.Sprintf("\n%s", vn))
	}
	return strings.Join(messages, ". ")
}

func (v *variables) merge(oth *variables) {
	for key, value := range oth.defs {
		v.defs[key] = value
	}
	for key, value := range oth.defsSuper {
		v.defsSuper[key] = value
	}
}

func (v *variables) addOnLetHookSetup(name VarID) {
	v.onLet[name] = struct{}{}
}

func (v *variables) tryRegisterVarBefore(name VarID) bool {
	if _, ok := v.before[name]; ok {
		return false
	}
	v.before[name] = struct{}{}
	return true
}

func (v *variables) hasOnLetHookApplied(name VarID) bool {
	_, ok := v.onLet[name]
	return ok
}

func (v *variables) rLock(varName VarID) func() {
	m := v.getMutex(varName)
	m.RLock()
	return m.RUnlock
}

func (v *variables) lock(varName VarID) func() {
	m := v.getMutex(varName)
	m.Lock()
	return m.Unlock
}

func (v *variables) getMutex(varName VarID) *sync.RWMutex {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if _, ok := v.locks[varName]; !ok {
		v.locks[varName] = &sync.RWMutex{}
	}
	return v.locks[varName]
}

//////////////////////////////////////////////////////// super /////////////////////////////////////////////////////////

func (v *variables) SetSuper(varName VarID, val any) {
	v.cacheSuper.Set(varName, val)
}

func (v *variables) LookupSuper(t *T, varName VarID) (any, bool) {
	if cv, ok := v.cacheSuper.Lookup(varName); ok {
		return cv, ok
	}
	var declOfSuper func(*T) any
	if decl, ok := v.cacheSuper.FindDecl(varName, v.defsSuper[varName]); ok {
		declOfSuper = decl
	}
	if declOfSuper == nil {
		return nil, false
	}
	stepOut := v.cacheSuper.StepIn(varName)
	val := declOfSuper(t)
	stepOut()
	v.SetSuper(varName, val)
	return val, true
}

func (v *variables) depsInitDo(id VarID, fn func()) {
	v.depsInitFor(id).Do(fn)
}

func (v *variables) depsInitFor(id VarID) *sync.Once {
	//
	// FAST
	v.mutex.RLock()
	once, ok := v.deps[id]
	v.mutex.RUnlock()
	if ok && once != nil {
		return once
	}
	//
	// SLOW
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if _, ok := v.deps[id]; !ok {
		v.deps[id] = &sync.Once{}
	}
	return v.deps[id]
}

func newVariablesSuperCache() *variablesSuperCache {
	return &variablesSuperCache{
		cache:        make(map[VarID]map[int]any),
		currentDepth: make(map[VarID]int),
	}
}

type variablesSuperCache struct {
	cache        map[VarID]map[int]any
	currentDepth map[VarID]int
}

func (sc *variablesSuperCache) StepIn(varName VarID) func() {
	if sc.currentDepth == nil {
		sc.currentDepth = make(map[VarID]int)
	}
	sc.currentDepth[varName]++
	return func() { sc.currentDepth[varName]-- }
}

func (sc *variablesSuperCache) depthFor(varName VarID) int {
	if sc.currentDepth == nil {
		return 0
	}
	return sc.currentDepth[varName]
}

func (sc *variablesSuperCache) Lookup(varName VarID) (any, bool) {
	if sc.cache == nil {
		return nil, false
	}
	dvs, ok := sc.cache[varName]
	if !ok {
		return nil, false
	}
	v, ok := dvs[sc.depthFor(varName)]
	return v, ok
}

func (sc *variablesSuperCache) Set(varName VarID, v any) {
	if sc.cache == nil {
		sc.cache = make(map[VarID]map[int]any)
	}
	if _, ok := sc.cache[varName]; !ok {
		sc.cache[varName] = make(map[int]any)
	}
	sc.cache[varName][sc.depthFor(varName)] = v
}

func (sc *variablesSuperCache) FindDecl(varName VarID, defs []variablesInitBlock) (variablesInitBlock, bool) {
	if defs == nil {
		return nil, false
	}
	depthIndex := sc.depthFor(varName)
	if !(depthIndex < len(defs)) {
		return nil, false
	}
	return defs[depthIndex], true
}
