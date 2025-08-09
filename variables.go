package testcase

import (
	"fmt"
	"strings"
	"sync"

	"go.llib.dev/testcase/internal/slicekit"
)

func newVariables() *variables {
	return &variables{
		defs:   make(map[VarID][]variablesInitBlock),
		cache:  make(map[vsk]any),
		depth:  make(variablesDepth),
		onLet:  make(map[VarID]struct{}),
		locks:  make(map[vsk]*sync.RWMutex),
		before: make(map[VarID]struct{}),
		deps:   make(map[VarID]*sync.Once),
	}
}

type vsk struct {
	VarID
	Depth int
}

// variables represents an individual test case's runtime variables.
// Using the variables cache within the individual test cases are safe even with *testing#T.Parallel().
// Different test cases don't share they variables instance.
type variables struct {
	mutex  sync.RWMutex
	cache  map[vsk]any
	locks  map[vsk]*sync.RWMutex
	depth  variablesDepth
	defs   map[VarID][]variablesInitBlock
	onLet  map[VarID]struct{}
	before map[VarID]struct{}
	deps   map[VarID]*sync.Once
}

type variablesInitBlock func(t *T) any

type variablesDepth map[VarID]int

func (m *variables) depthStep(id VarID) func() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.depth.add(id, 1)
	return func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.depth.add(id, -1)
	}
}

func (m *variablesDepth) add(id VarID, n int) {
	if *m == nil {
		*m = make(variablesDepth)
	}
	(*m)[id] = (*m)[id] + n
}

func (m *variablesDepth) Inc(id VarID) { m.add(id, 1) }
func (m *variablesDepth) Dec(id VarID) { m.add(id, -1) }

func (m variablesDepth) Get(id VarID) int {
	if m == nil {
		return 0
	}
	d, ok := m[id]
	if !ok {
		return 0
	}
	return d
}

func (vs *variables) lookupCache(id VarID) (any, bool) {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()
	if vs.cache == nil {
		return nil, false
	}
	val, ok := vs.cache[vs.key(id)]
	return val, ok
}

func (vs *variables) setCache(id VarID, v any) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	if vs.cache == nil {
		vs.cache = make(map[vsk]any)
	}
	vs.cache[vs.key(id)] = v
}

func (vs *variables) lookupDef(varName VarID) (variablesInitBlock, bool) {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()
	if vs.defs == nil {
		return nil, false
	}
	defs, ok := vs.defs[varName]
	if !ok {
		return nil, false
	}
	depth := vs.depth.Get(varName)
	return slicekit.ReverseLookup(defs, depth)
}

func (vs *variables) Knows(varName VarID) bool {
	m := vs.getMutex(varName)
	m.RLock()
	defer m.RUnlock()
	if _, ok := vs.lookupDef(varName); ok {
		return true
	}
	if _, ok := vs.lookupCache(varName); ok {
		return true
	}
	return false
}

func (vs *variables) Let(id VarID, blk variablesInitBlock /* [any] */) {
	m := vs.getMutex(id)
	m.Lock()
	defer m.Unlock()
	vs.let(id, blk)
}

func (vs *variables) let(varName VarID, blk variablesInitBlock /* [any] */) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	vs.defs[varName] = append(vs.defs[varName], blk)
}

func (vs *variables) Get(t *T, id VarID) any {
	helper(t.TB).Helper()
	v, ok := vs.Lookup(t, id)
	if !ok {
		t.Fatal(vs.fatalMessageFor(id))
	}
	return v
}

// Get will return a testcase vs.
//
// If there is no such value, then it will panic with a "friendly" message.
func (vs *variables) Lookup(t *T, id VarID) (any, bool) {
	helper(t.TB).Helper()

	m := vs.getMutex(id)

	m.RLock()
	v, ok := vs.lookupCache(id)
	m.RUnlock()

	if ok {
		return v, true
	}

	m.Lock()
	defer m.Unlock()

	if v, ok := vs.lookupCache(id); ok {
		return v, true
	}
	def, ok := vs.lookupDef(id)
	if !ok || def == nil {
		return nil, false
	}

	v = def(t)
	vs.setCache(id, v)
	return v, true
}

func (vs *variables) key(id VarID) vsk {
	return vsk{
		VarID: id,
		Depth: vs.depth.Get(id),
	}
}

func (vs *variables) cacheSet(id VarID, data any) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	vs.cache[vs.key(id)] = data
}

func (vs *variables) Set(id VarID, value any) {
	m := vs.getMutex(id)
	m.Lock()
	defer m.Unlock()
	vs.cacheSet(id, value)
}

func (vs *variables) reset() {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	vs.cache = make(map[vsk]any)
}

func (vs *variables) fatalMessageFor(id VarID) string {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()
	var messages []string
	messages = append(messages, fmt.Sprintf(`Variable %q is not found`, id))
	var ids []VarID
	for id := range vs.defs {
		ids = append(ids, id)
	}
	messages = append(messages, `Did you mean?`)
	for _, vn := range ids {
		messages = append(messages, fmt.Sprintf("\n%s", vn))
	}
	return strings.Join(messages, ". ")
}

func (vs *variables) merge(oth *variables) {
	for id, ds := range oth.defs {
		vs.defs[id] = append(vs.defs[id], ds...)
	}
	// all the other fields is basically runtime states
	// we don't need to merge those as part of merge
	// we only care about the definitions.
}

func (vs *variables) addOnLetHookSetup(name VarID) {
	vs.onLet[name] = struct{}{}
}

func (vs *variables) tryRegisterVarBefore(name VarID) bool {
	if _, ok := vs.before[name]; ok {
		return false
	}
	vs.before[name] = struct{}{}
	return true
}

func (vs *variables) hasOnLetHookApplied(name VarID) bool {
	_, ok := vs.onLet[name]
	return ok
}

func (vs *variables) getMutex(id VarID) *sync.RWMutex {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	var key = vs.key(id)
	if _, ok := vs.locks[key]; !ok {
		vs.locks[key] = &sync.RWMutex{}
	}
	return vs.locks[key]
}

//////////////////////////////////////////////////////// super /////////////////////////////////////////////////////////

func (vs *variables) SetSuper(id VarID, val any) {
	defer vs.depthStep(id)()
	vs.Set(id, val)
}

func (vs *variables) LookupSuper(t *T, id VarID) (any, bool) {
	defer vs.depthStep(id)()
	return vs.Lookup(t, id)
}

func (vs *variables) depsInitDo(id VarID, fn func()) {
	vs.depsInitFor(id).Do(fn)
}

func (vs *variables) depsInitFor(id VarID) *sync.Once {
	//
	// FAST
	vs.mutex.RLock()
	once, ok := vs.deps[id]
	vs.mutex.RUnlock()
	if ok && once != nil {
		return once
	}
	//
	// SLOW
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	if _, ok := vs.deps[id]; !ok {
		vs.deps[id] = &sync.Once{}
	}
	return vs.deps[id]
}
