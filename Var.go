package testcase

import (
	"fmt"
	"reflect"
)

// Var is a testCase helper structure, that allows easy way to access testCase runtime variables.
// In the future it will be updated to use Go2 type parameters.
//
// Var allows creating testCase variables in a modular way.
// By modular, imagine that you can have commonly used values initialized and then access it from the testCase runtime spec.
// This approach allows an easy dependency injection maintenance at project level for your testing suite.
// It also allows you to have parallel testCase execution where you don't expect side effect from your subject.
//
//	e.g.: HTTP JSON API testCase and GraphQL testCase both use the business rule instances.
//	Or multiple business rules use the same storage dependency.
//
// The last use-case it allows is to define dependencies for your testCase subject before actually assigning values to it.
// Then you can focus on building up the testing spec and assign values to the variables at the right testing subcontext. With variables, it is easy to forget to assign a value to a variable or forgot to clean up the value of the previous run and then scratch the head during debugging.
// If you forgot to set a value to the variable in testcase, it warns you that this value is not yet defined to the current testing scope.
type Var[V any] struct {
	// ID is the testCase spec variable group from where the cached value can be accessed later on.
	// ID is Mandatory when you create a variable, else the empty string will be used as the variable group.
	ID string
	// Init is an optional constructor definition that will be used when Var is bonded to a *Spec without constructor function passed to the Let function.
	// The goal of this field to initialize a variable that can be reused across different testing suites by bounding the Var to a given testing suite.
	//
	// Please use #Get if you wish to access a testCase runtime across cached variable value.
	// The value returned by this is not subject to any #Before and #Around hook that might mutate the variable value during the testCase runtime.
	// Init function doesn't cache the value in the testCase runtime spec but literally just meant to initialize a value for the Var in a given test case.
	// Please use it with caution.
	Init VarInit[V]
	// Before is a hook that will be executed once during the lifetime of tests that uses the Var.
	// If the Var is not bound to the Spec at Spec.Context level, the Before Hook will be executed at Var.Get.
	Before func(t *T, v Var[V])
	// OnLet is an optional Var hook that is executed when the variable being bind to Spec context.
	// This hook is ideal to set up tags on the Spec, call Spec.Sequential
	// or ensure binding of further dependencies that this variable requires.
	//
	// In case OnLet is provided, the Var must be explicitly set to a Spec with a Let call
	// else accessing the Var value will panic and warn about this.
	OnLet func(s *Spec, v Var[V])
}

type VarInit[V any] func(*T) V

//func CastToVarInit[V any](fn func(testing.TB) V) func(*T) V {
//	if fn == nil {
//		return nil
//	}
//	return func(t *T) V { return fn(t) }
//}

const (
	varOnLetNotInitialized = `%s Var has Var.OnLet. You must use Var.Let, Var.LetValue to initialize it properly.`
	varIDIsIsMissing       = `ID for %T is missing. Maybe it's uninitialized?`
)

// Get returns the current cached value of the given Variable
// Get is a thread safe operation.
// When Go2 released, it will replace type casting
func (v Var[V]) Get(t *T) V {
	t.Helper()
	defer t.pauseTimer()()
	if v.ID == "" {
		t.Fatalf(varIDIsIsMissing, v)
	}
	if v.OnLet != nil && !t.hasOnLetHookApplied(v.ID) {
		t.Fatalf(varOnLetNotInitialized, v.ID)
	}
	v.execBefore(t)
	if !t.vars.Knows(v.ID) && v.Init != nil {
		t.vars.Let(v.ID, func(t *T) interface{} { return v.Init(t) })
	}
	rv, ok := t.vars.Get(t, v.ID).(V)
	if !ok && t.vars.Get(t, v.ID) != nil {
		t.Logf("Incorrect value type for Var.ID: %q", v.ID)
		t.Log("If you use .Var type without the .Let helper method")
		t.Log("then please make sure that the Var.ID field is unique between your Var instances.")
		t.Logf("expected: %T", *new(V))
		t.Logf("actual: %T", t.vars.Get(t, v.ID))
		t.FailNow()
	}
	return rv
}

// Set sets a value to a given variable during testCase runtime
// Set is a thread safe operation.
func (v Var[V]) Set(t *T, value V) {
	t.Helper()
	if v.OnLet != nil && !t.hasOnLetHookApplied(v.ID) {
		t.Fatalf(varOnLetNotInitialized, v.ID)
	}
	t.vars.Set(v.ID, value)
}

// Let allow you to set the variable value to a given spec
func (v Var[V]) Let(s *Spec, blk VarInit[V]) Var[V] {
	s.testingTB.Helper()
	v.onLet(s)
	return let(s, v.ID, blk)
}

type letWithSuperBlock[V any] func(t *T, super V) V

func (v Var[V]) onLet(s *Spec) {
	s.testingTB.Helper()
	if v.OnLet != nil {
		v.OnLet(s, v)
		s.vars.addOnLetHookSetup(v.ID)
	}
	if v.Before != nil {
		s.Before(v.execBefore)
	}
}

func (v Var[V]) execBefore(t *T) {
	t.Helper()
	if v.Before != nil && t.vars.tryRegisterVarBefore(v.ID) {
		v.Before(t, v)
	}
}

// LetValue set the value of the variable to a given block
func (v Var[V]) LetValue(s *Spec, value V) Var[V] {
	s.testingTB.Helper()
	v.onLet(s)
	return letValue[V](s, v.ID, value)
}

// Bind is a syntax sugar shorthand for Var.Let(*Spec, nil),
// where skipping providing a block meant to be explicitly expressed.
func (v Var[V]) Bind(s *Spec) Var[V] {
	s.testingTB.Helper()
	for _, s := range s.specsFromCurrent() {
		if s.vars.Knows(v.ID) {
			return v
		}
	}
	return v.Let(s, v.Init)
}

// EagerLoading allows the variable to be loaded before the action and assertion block is reached.
// This can be useful when you want to have a variable that cause side effect on your system.
// Like it should be present in some sort of attached resource/storage.
//
// For example, you may persist the value in a storage as part of the initialization block,
// and then when the testCase/then block is reached, the entity is already present in the resource.
func (v Var[V]) EagerLoading(s *Spec) Var[V] {
	s.testingTB.Helper()
	s.Before(func(t *T) { _ = v.Get(t) })
	return v
}

// Super will return the inherited Super value of your Var.
// This means that if you declared Var in an outer Spec.Context,
// or your Var has an Var.Init field, then Var.Super will return its content.
// This allows you to incrementally extend with values the inherited value until you reach your testing scope.
// This also allows you to wrap your Super value with a Spy or Stub wrapping layer,
// and pry the interactions with the object while using the original value as a base.
func (v Var[V]) Super(t *T) V {
	t.Helper()
	isuper, ok := t.vars.LookupSuper(t, v.ID)
	if !ok && v.Init != nil {
		isuper = any(v.Init(t))
		ok = true
		t.vars.SetSuper(v.ID, isuper)
	}
	if !ok {
		panic(fmt.Sprintf("no super/previous value decleration found for Var[%T]. Are you sure you defined one already?", *new(V)))
	}
	return isuper.(V)
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
