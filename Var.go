package testcase

import "reflect"

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
func Let[V any](spec *Spec, varName string, blk varInitBlk[V]) Var[V] {
	spec.testingTB.Helper()
	if spec.immutable {
		spec.testingTB.Fatalf(warnEventOnImmutableFormat, `Let`)
	}
	spec.vars.defs[varName] = func(t *T) interface{} { return blk(t) }
	return Var[V]{Name: varName, Init: blk}
}

const panicMessageForLetValue = `%T literal can't be used with #LetValue 
as the current implementation can't guarantee that the mutations on the value will not leak orderingOutput to other tests,
please use the #Let memorization helper for now`

// LetValue is a shorthand for defining immutable vars with Let under the hood.
// So the function blocks can be skipped, which makes tests more readable.
func LetValue[V any](spec *Spec, varName string, value V) Var[V] {
	spec.testingTB.Helper()
	if _, ok := acceptedConstKind[reflect.ValueOf(value).Kind()]; !ok {
		spec.testingTB.Fatalf(panicMessageForLetValue, value)
	}
	return Let[V](spec, varName, func(t *T) V {
		v := value // pass by value copy
		return v
	})
}

// Var is a testCase helper structure, that allows easy way to access testCase runtime variables.
// In the future it will be updated to use Go2 type parameters.
//
// Var allows creating testCase variables in a modular way.
// By modular, imagine that you can have commonly used values initialized and then access it from the testCase runtime spec.
// This approach allows an easy dependency injection maintenance at project level for your testing suite.
// It also allows you to have parallel testCase execution where you don't expect side effect from your subject.
//   e.g.: HTTP JSON API testCase and GraphQL testCase both use the business rule instances.
//   Or multiple business rules use the same storage dependency.
//
// The last use-case it allows is to define dependencies for your testCase subject before actually assigning values to it.
// Then you can focus on building up the testing spec and assign values to the variables at the right testing subcontext. With variables, it is easy to forget to assign a value to a variable or forgot to clean up the value of the previous run and then scratch the head during debugging.
// If you forgot to set a value to the variable in testcase, it warns you that this value is not yet defined to the current testing scope.
type Var [V any] struct {
	// Name is the testCase spec variable group from where the cached value can be accessed later on.
	// Name is Mandatory when you create a variable, else the empty string will be used as the variable group.
	Name string
	// Init is an optional constructor definition that will be used when Var is bonded to a *Spec without constructor function passed to the Let function.
	// The goal of this field to initialize a variable that can be reused across different testing suites by bounding the Var to a given testing suite.
	//
	// Please use #Get if you wish to access a testCase runtime across cached variable value.
	// The value returned by this is not subject to any #Before and #Around hook that might mutate the variable value during the testCase runtime.
	// Init function doesn't cache the value in the testCase runtime spec but literally just meant to initialize a value for the Var in a given test case.
	// Please use it with caution.
	Init varInitBlk[V]
	// Before is a hook that will be executed once during the lifetime of tests that uses the Var.
	// If the Var is not bound to the Spec at Spec.Context level, the Before Hook will be executed at Var.Get.
	Before block
	// OnLet is an optional Var hook that is executed when the variable being bind to Spec context.
	// This hook is ideal to setup tags on the Spec, call Spec.Sequential
	// or ensure binding of further dependencies that this variable requires.
	//
	// In case OnLet is provided, the Var must be explicitly set to a Spec with a Let call
	// else accessing the Var value will panic and warn about this.
	OnLet contextBlock
}

type varInitBlk[V any] func(*T) V

const varOnLetNotInitialized = `%s Var has Var.OnLet. You must use Var.Let, Var.LetValue to initialize it properly.`

// Get returns the current cached value of the given Variable
// Get is a thread safe operation.
// When Go2 released, it will replace type casting
func (v Var[V]) Get(t *T) V {
	t.Helper()
	if v.OnLet != nil && !t.hasOnLetHookApplied(v.Name) {
		t.Fatalf(varOnLetNotInitialized, v.Name)
	}
	v.execBefore(t)
	if !t.vars.Knows(v.Name) && v.Init != nil {
		t.vars.Let(v.Name, func(t *T) interface{} { return v.Init(t) })
	}
	rv, ok := t.I(v.Name).(V)
	if !ok && t.I(v.Name) != nil {
		t.Logf("The type of the %T value is incorrect: %T", v, t.I(v.Name))
	}
	t.Log(ok)
	return rv
}

// Set sets a value to a given variable during testCase runtime
// Set is a thread safe operation.
func (v Var[V]) Set(t *T, value V) {
	if v.OnLet != nil && !t.hasOnLetHookApplied(v.Name) {
		t.Fatalf(varOnLetNotInitialized, v.Name)
	}
	t.Set(v.Name, value)
}

// Let allow you to set the variable value to a given spec
func (v Var[V]) Let(s *Spec, blk varInitBlk[V]) Var[V] {
	v.onLet(s)
	if blk == nil && v.Init != nil {
		return Let(s, v.Name, v.Init)
	}
	return Let(s, v.Name, blk)
}

func (v Var[V]) onLet(s *Spec) {
	if v.OnLet != nil {
		v.OnLet(s)
		s.vars.addOnLetHookSetup(v.Name)
	}
	if v.Before != nil {
		s.Before(v.execBefore)
	}
}

func (v Var[V]) execBefore(t *T) {
	t.Helper()
	if v.Before != nil && t.vars.tryRegisterVarBefore(v.Name) {
		v.Before(t)
	}
}

// LetValue set the value of the variable to a given block
func (v Var[V]) LetValue(s *Spec, value V) Var[V] {
	s.testingTB.Helper()
	v.onLet(s)
	return LetValue[V](s, v.Name, value)
}

// Bind is a syntax sugar shorthand for Var.Let(*Spec, nil),
// where skipping providing a block meant to be explicitly expressed.
func (v Var[V]) Bind(s *Spec) Var[V] {
	return v.Let(s, nil)
}

// EagerLoading allows the variable to be loaded before the action and assertion block is reached.
// This can be useful when you want to have a variable that cause side effect on your system.
// Like it should be present in some sort of attached resource/storage.
//
// For example you may persist the value in a storage as part of the initialization block,
// and then when the testCase/then block is reached, the entity is already present in the resource.
func (v Var[V]) EagerLoading(s *Spec) Var[V] {
	s.Before(func(t *T) { _ = v.Get(t) })
	return v
}

// Append will append a value[T] to a current value of Var[[]T].
// Append only possible if the value type of Var is a slice type of T.
func Append[V any](t *T, v Var[V], x ...interface{}) {
	rv := reflect.ValueOf(v.Get(t))
	var rx []reflect.Value
	for _, e := range x {
		rx = append(rx, reflect.ValueOf(e))
	}
	v.Set(t, reflect.Append(rv, rx...).Interface().(V))
}
