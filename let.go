package testcase

import (
	"fmt"
	"reflect"

	"github.com/adamluzsi/testcase/internal"
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
//
func Let[V any](spec *Spec, blk varInitBlk[V]) Var[V] {
	spec.testingTB.Helper()
	return let[V](spec, makeVarName(spec), blk)
}

const panicMessageForLetValue = `%T literal can't be used with #LetValue 
as the current implementation can't guarantee that the mutations on the value will not leak orderingOutput to other tests,
please use the #Let memorization helper for now`

// LetValue is a shorthand for defining immutable vars with Let under the hood.
// So the function blocks can be skipped, which makes tests more readable.
func LetValue[V any](spec *Spec, value V) Var[V] {
	spec.testingTB.Helper()
	return letValue[V](spec, makeVarName(spec), value)
}

func let[V any](spec *Spec, varName string, blk varInitBlk[V]) Var[V] {
	spec.testingTB.Helper()
	if spec.immutable {
		spec.testingTB.Fatalf(warnEventOnImmutableFormat, `Let`)
	}
	if blk != nil {
		spec.vars.defs[varName] = func(t *T) interface{} { return blk(t) }
	}
	return Var[V]{ID: varName, Init: blk}
}

func letValue[V any](spec *Spec, varName string, value V) Var[V] {
	spec.testingTB.Helper()
	if _, ok := acceptedConstKind[reflect.ValueOf(value).Kind()]; !ok {
		spec.testingTB.Fatalf(panicMessageForLetValue, value)
	}
	return let[V](spec, varName, func(t *T) V {
		v := value // pass by value copy
		return v
	})
}

func makeVarName(spec *Spec) string {
	spec.testingTB.Helper()
	location := internal.CallerLocation(1, false)
	// when variable is declared within a loop
	// providing a variable ID offset is required to identify the variable uniquely.

	varNameIndex := make(map[string]struct{})
	for _, s := range spec.list() {
		for k := range s.vars.defs {
			varNameIndex[k] = struct{}{}
		}
	}

	var (
		name   string
		offset int
	)
positioning:
	for {
		// quick path for the majority of the case.
		if _, ok := varNameIndex[location]; !ok {
			name = location
			break positioning
		}

		offset++
		name = fmt.Sprintf("%s#[%d]", location, offset)
		if _, ok := varNameIndex[name]; !ok {
			break positioning
		}
	}
	return name
}
