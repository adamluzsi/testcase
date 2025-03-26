package testcase

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"go.llib.dev/testcase/internal/caller"
	"go.llib.dev/testcase/internal/reflects"
)

// Let define a memoized helper method.
// Let creates lazily-evaluated test execution bound variables.
// Let variables don't exist until called into existence by the actual tests,
// so you won't waste time loading them for examples that don't use them.
// They're also memoized, so they're useful for encapsulating database objects, due to the cost of making a database request.
// The value will be cached across list use within the same test execution but not across different test cases.
// You can eager load a value defined in let by referencing to it in a Before hook.
// Let is threadsafe, the parallel running test will receive they own test variable instance.
//
// Defining a value in a spec Context will ensure that the scope
// and it's nested scopes of the current scope will have access to the value.
// It cannot leak its value outside from the current scope.
// Calling Let in a nested/sub scope will apply the new value for that value to that scope and below.
//
// It will panic if it is used after a When/And/Then scope definition,
// because those scopes would have no clue about the later defined variable.
// In order to keep the specification reading mental model requirement low,
// it is intentionally not implemented to handle such case.
// Defining test vars always expected in the beginning of a specification scope,
// mainly for readability reasons.
//
// vars strictly belong to a given `Describe`/`When`/`And` scope,
// and configured before any hook would be applied,
// therefore hooks always receive the most latest version from the `Let` vars,
// regardless in which scope the hook that use the variable is define.
//
// Let can enhance readability
// when used sparingly in any given example group,
// but that can quickly degrade with heavy overuse.
func Let[V any](spec *Spec, blk VarInit[V]) Var[V] {
	helper(spec.testingTB).Helper()
	return let[V](spec, makeVarID(spec), blk)
}

type tuple2[V, B any] struct {
	V V
	B B
}

// Let2 is a tuple-style variable creation method, where an init block is shared between different variables.
func Let2[V, B any](spec *Spec, blk func(*T) (V, B)) (Var[V], Var[B]) {
	helper(spec.testingTB).Helper()
	src := Let[tuple2[V, B]](spec, func(t *T) tuple2[V, B] {
		v, b := blk(t)
		return tuple2[V, B]{V: v, B: b}
	})
	return Let[V](spec, func(t *T) V {
			return src.Get(t).V
		}), Let[B](spec, func(t *T) B {
			return src.Get(t).B
		})
}

type tuple3[V, B, N any] struct {
	V V
	B B
	N N
}

// Let3 is a tuple-style variable creation method, where an init block is shared between different variables.
func Let3[V, B, N any](spec *Spec, blk func(*T) (V, B, N)) (Var[V], Var[B], Var[N]) {
	helper(spec.testingTB).Helper()
	src := Let[tuple3[V, B, N]](spec, func(t *T) tuple3[V, B, N] {
		v, b, n := blk(t)
		return tuple3[V, B, N]{V: v, B: b, N: n}
	})
	return Let[V](spec, func(t *T) V {
			return src.Get(t).V
		}), Let[B](spec, func(t *T) B {
			return src.Get(t).B
		}), Let[N](spec, func(t *T) N {
			return src.Get(t).N
		})
}

const panicMessageForLetValue = `%T literal can't be used with #LetValue 
as the current implementation can't guarantee that the mutations on the value will not leak orderingOutput to other tests,
please use the #Let memorization helper for now`

// LetValue is a shorthand for defining immutable vars with Let under the hood.
// So the function blocks can be skipped, which makes tests more readable.
func LetValue[V any](spec *Spec, value V) Var[V] {
	helper(spec.testingTB).Helper()
	return letValue[V](spec, makeVarID(spec), value)
}

func let[V any](spec *Spec, varID VarID, blk VarInit[V]) Var[V] {
	helper(spec.testingTB).Helper()
	if spec.immutable {
		spec.testingTB.Fatalf(warnEventOnImmutableFormat, `Let`)
	}
	if blk != nil {
		spec.vars.defsSuper[varID] = findCurrentDeclsFor(spec, varID)
		spec.vars.defs[varID] = func(t *T) any {
			t.Helper()
			return blk(t)
		}
	}
	return Var[V]{ID: varID, Init: blk}
}

func letValue[V any](spec *Spec, varName VarID, value V) Var[V] {
	helper(spec.testingTB).Helper()
	if isMutable[V](value) {
		spec.testingTB.Fatalf(panicMessageForLetValue, value)
	}
	return let[V](spec, varName, func(t *T) V {
		t.Helper()
		v := value // pass by value copy
		return v
	})
}

// latest decl is the first and the deeper you want to reach back, the higher the index
func findCurrentDeclsFor(spec *Spec, varName VarID) []variablesInitBlock {
	var decls []variablesInitBlock
	for _, s := range spec.specsFromCurrent() {
		if decl, ok := s.vars.defs[varName]; ok {
			decls = append(decls, decl)
		}
	}
	return decls
}

func makeVarID(spec *Spec) VarID {
	helper(spec.testingTB).Helper()
	location := caller.GetLocation(false)
	// when variable is declared within a loop
	// providing a variable ID offset is required to identify the variable uniquely.

	varIDIndex := make(map[VarID]struct{})
	for _, s := range spec.specsFromParent() {
		for k := range s.vars.locks {
			varIDIndex[k] = struct{}{}
		}
		for k := range s.vars.defs {
			varIDIndex[k] = struct{}{}
		}
		for k := range s.vars.onLet {
			varIDIndex[k] = struct{}{}
		}
		for k := range s.vars.before {
			varIDIndex[k] = struct{}{}
		}
	}

	var (
		id     VarID
		offset int
	)
positioning:
	for {
		// quick path for the majority of the case.
		if _, ok := varIDIndex[VarID(location)]; !ok {
			id = VarID(location)
			break positioning
		}

		offset++
		id = VarID(fmt.Sprintf("%s#[%d]", location, offset))
		if _, ok := varIDIndex[id]; !ok {
			break positioning
		}
	}
	return id
}

var mutableException = map[reflect.Type]struct{}{}

// RegisterImmutableType In some cases, certain types are actually immutable, but use a mutable type to represent that immutable value type.
// For example, time.Location is such case.
func RegisterImmutableType[T any]() func() {
	rtype := reflect.TypeOf((*T)(nil)).Elem()
	mutableException[rtype] = struct{}{}
	return func() { delete(mutableException, rtype) }
}

func isMutable[T any](v T) bool {
	rtype := reflect.TypeOf((*T)(nil)).Elem()
	if _, ok := mutableException[rtype]; ok {
		return false
	}
	if _, ok := mutableException[reflect.TypeOf(v)]; ok {
		return false
	}
	return reflects.IsMutable(v)
}

var _ = RegisterImmutableType[time.Time]()
var _ = RegisterImmutableType[time.Location]()
var _ = RegisterImmutableType[reflect.Type]()
var _ = RegisterImmutableType[regexp.Regexp]()
