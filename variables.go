package testcase

import (
	"fmt"
	"strings"
	"sync"

	"go.llib.dev/testcase/internal/slicekit"
	"go.llib.dev/testcase/pp"
)

func newVariables() *variables {
	return &variables{
		definitions: make(map[VarID][]variablesInitBlock),
		cachev:      make(map[cacheK]any),
		depth:       make(variablesDepth),
		onLet:       make(map[VarID]struct{}),
		locks:       make(map[VarID]*sync.RWMutex),
		before:      make(map[VarID]struct{}),
		deps:        make(map[VarID]*sync.Once),
		// TODO: deprecate below
		//
		defs:       make(map[VarID]variablesInitBlock),
		cached:     make(map[VarID]any),
		defsSuper:  make(map[VarID][]variablesInitBlock),
		cacheSuper: newVariablesSuperCache(),
	}
}

type cacheK struct {
	VarID
	Depth int
}

// variables represents an individual test case's runtime variables.
// Using the variables cache within the individual test cases are safe even with *testing#T.Parallel().
// Different test cases don't share they variables instance.
type variables struct {
	mutex       sync.RWMutex
	locks       map[VarID]*sync.RWMutex
	defs        map[VarID]variablesInitBlock
	definitions map[VarID][]variablesInitBlock
	defsSuper   map[VarID][]variablesInitBlock
	onLet       map[VarID]struct{}
	before      map[VarID]struct{}
	cachev      map[cacheK]any
	depth       variablesDepth

	cached     map[VarID]any
	cacheSuper *variablesSuperCache
	deps       map[VarID]*sync.Once
}

type variablesInitBlock func(t *T) any

type variablesDepth map[VarID]int

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
	if vs.cachev == nil {
		return nil, false
	}
	val, ok := vs.cachev[vs.cacheKey(id)]
	return val, ok
}

func (vs *variables) lookupDef(varName VarID) (variablesInitBlock, bool) {
	if vs.definitions == nil {
		return nil, false
	}
	defs, ok := vs.definitions[varName]
	if !ok {
		return nil, false
	}
	depth := vs.depth.Get(varName)
	return slicekit.Lookup(defs, -1*depth)
}

func (vs *variables) Knows(varName VarID) bool {
	defer vs.rLock(varName)()
	if _, ok := vs.lookupDef(varName); ok {
		return true
	}
	if _, ok := vs.lookupCache(varName); ok {
		return true
	}
	return false
}

func (vs *variables) Let(varName VarID, blk variablesInitBlock /* [any] */) {
	defer vs.lock(varName)()
	vs.let(varName, blk)
}

func (vs *variables) let(varName VarID, blk variablesInitBlock /* [any] */) {
	vs.definitions[varName] = append(vs.definitions[varName], blk)
	vs.defs[varName] = blk
}

// Get will return a testcase vs.
//
// If there is no such value, then it will panic with a "friendly" message.
func (vs *variables) Get(t *T, varName VarID) any {
	t.TB.Helper()
	if !vs.Knows(varName) {
		t.Fatal(vs.fatalMessageFor(varName))
	}
	defer vs.lock(varName)()
	if !vs.cacheHas(varName) {
		// cacheSet(varName, ...) is protected from concurrent access by lock(varName).
		vs.cacheSet(varName, vs.defs[varName](t))
	}
	return t.vars.cacheGet(varName)
}

func (vs *variables) cacheKey(id VarID) cacheK {
	return cacheK{
		VarID: id,
		Depth: vs.depth.Get(id),
	}
}

func (vs *variables) cacheGet(varName VarID) any {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()
	return vs.cached[varName]
}

func (vs *variables) cacheHas(varName VarID) bool {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()
	if _, ok := vs.cachev[vs.cacheKey(varName)]; ok {
		return true
	}
	if _, ok := vs.cached[varName]; ok {
		return true
	}
	return false
}

func (vs *variables) cacheSet(id VarID, data any) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	vs.cached[id] = data
	vs.cachev[vs.cacheKey(id)] = data
}

func (vs *variables) Set(varName VarID, value any) {
	defer vs.lock(varName)()
	if _, ok := vs.defs[varName]; !ok {
		vs.let(varName, func(t *T) any { return value })
	}
	vs.cacheSet(varName, value)
}

func (vs *variables) reset() {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	vs.cached = make(map[VarID]any)
	vs.cachev = make(map[cacheK]any)
	vs.cacheSuper = newVariablesSuperCache()
}

func (vs *variables) fatalMessageFor(varName VarID) string {
	var messages []string
	messages = append(messages, fmt.Sprintf(`Variable %q is not found`, varName))
	var keys []VarID
	for k := range vs.defs {
		keys = append(keys, k)
	}
	messages = append(messages, `Did you mean?`)
	for _, vn := range keys {
		messages = append(messages, fmt.Sprintf("\n%s", vn))
	}
	return strings.Join(messages, ". ")
}

func (vs *variables) merge(oth *variables) {
	for key, value := range oth.defs {
		vs.defs[key] = value
	}
	for key, value := range oth.defsSuper {
		vs.defsSuper[key] = value
	}
	for vn, ds := range oth.definitions {
		vs.definitions[vn] = append(vs.definitions[vn], ds...)
	}
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

func (vs *variables) rLock(varName VarID) func() {
	m := vs.getMutex(varName)
	m.RLock()
	return m.RUnlock
}

func (vs *variables) lock(varName VarID) func() {
	m := vs.getMutex(varName)
	m.Lock()
	return m.Unlock
}

func (vs *variables) getMutex(varName VarID) *sync.RWMutex {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	if _, ok := vs.locks[varName]; !ok {
		vs.locks[varName] = &sync.RWMutex{}
	}
	return vs.locks[varName]
}

//////////////////////////////////////////////////////// super /////////////////////////////////////////////////////////

func (vs *variables) setSuper(varName VarID, val any) {
	vs.cacheSuper.Set(varName, val)
	vs.cachev[vs.cacheKey(varName)] = val
}

func (vs *variables) SetSuper(varName VarID, val any) {
	vs.depth.add(varName, 1)
	defer vs.depth.add(varName, -1)
	vs.setSuper(varName, val)
}

func (vs *variables) LookupSuper(t *T, varName VarID) (any, bool) {
	vs.depth.add(varName, 1)
	defer vs.depth.add(varName, -1)

	pp.PP("depth:", t.vars.depth.Get(varName))

	if v, ok := vs.lookupCache(varName); ok {
		return v, ok
	}
	if cv, ok := vs.cacheSuper.Lookup(varName); ok {
		return cv, ok
	}
	var declOfSuper func(*T) any

	if decl, ok := vs.cacheSuper.FindDecl(varName, vs.definitions[varName]); ok {
		// if decl, ok := v.cacheSuper.FindDecl(varName, v.defsSuper[varName]); ok {
		declOfSuper = decl
	}
	if declOfSuper == nil {
		return nil, false
	}
	stepOut := vs.cacheSuper.StepIn(varName)
	val := declOfSuper(t)
	stepOut()
	vs.setSuper(varName, val)
	return val, true
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

func (sc *variablesSuperCache) FindDecl(varName VarID, definitions []variablesInitBlock) (variablesInitBlock, bool) {
	if definitions == nil {
		return nil, false
	}
	depthOffset := sc.depthFor(varName)
	definitionsLen := len(definitions)
	if definitionsLen <= depthOffset {
		return nil, false
	}
	index := definitionsLen - 1 + -1*depthOffset - 1
	pp.PP("len", definitionsLen)
	pp.PP("index", index)
	pp.PP("last ind:", definitionsLen-1)
	pp.PP("offset", -1*depthOffset)
	pp.PP(definitionsLen, index)
	if index < 0 || !(index < definitionsLen) {
		return nil, false
	}
	return definitions[index], true
}
