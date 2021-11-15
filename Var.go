package testcase

import "reflect"

// TODO: update Ts to [T] when Go2 released

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
type Var struct /* [T] */ {
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
	Init letBlock /* [T] */
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

const varOnLetNotInitialized = `%s Var has Var.OnLet. You must use Var.Let, Var.LetValue to initialize it properly.`

// Get returns the current cached value of the given Variable
// Get is a thread safe operation.
// When Go2 released, it will replace type casting
func (v Var) Get(t *T) (T interface{}) {
	t.Helper()
	if v.OnLet != nil && !t.hasOnLetHookApplied(v.Name) {
		t.Fatalf(varOnLetNotInitialized, v.Name)
	}
	v.execBefore(t)
	if !t.vars.Knows(v.Name) && v.Init != nil {
		t.vars.Let(v.Name, v.Init)
	}
	r, _ := t.I(v.Name).(interface{}) // cast to T
	return r
}

// Set sets a value to a given variable during testCase runtime
// Set is a thread safe operation.
func (v Var) Set(t *T, value interface{}) {
	if v.OnLet != nil && !t.hasOnLetHookApplied(v.Name) {
		t.Fatalf(varOnLetNotInitialized, v.Name)
	}
	t.Set(v.Name, value)
}

// Let allow you to set the variable value to a given spec
func (v Var) Let(s *Spec, blk letBlock) Var {
	v.onLet(s)
	if blk == nil && v.Init != nil {
		return s.Let(v.Name, v.Init)
	}
	return s.Let(v.Name, blk)
}

func (v Var) onLet(s *Spec) {
	if v.OnLet != nil {
		v.OnLet(s)
		s.vars.addOnLetHookSetup(v.Name)
	}
	if v.Before != nil {
		s.Before(v.execBefore)
	}
}

func (v Var) execBefore(t *T) {
	t.Helper()
	if v.Before != nil && t.vars.tryRegisterVarBefore(v.Name) {
		v.Before(t)
	}
}

// LetValue set the value of the variable to a given block
func (v Var) LetValue(s *Spec, value interface{}) Var {
	v.onLet(s)
	return s.LetValue(v.Name, value)
}

// Bind is a syntax sugar shorthand for Var.Let(*Spec, nil),
// where skipping providing a block meant to be explicitly expressed.
func (v Var) Bind(s *Spec) Var {
	return v.Let(s, nil)
}

// EagerLoading allows the variable to be loaded before the action and assertion block is reached.
// This can be useful when you want to have a variable that cause side effect on your system.
// Like it should be present in some sort of attached resource/storage.
//
// For example you may persist the value in a storage as part of the initialization block,
// and then when the testCase/then block is reached, the entity is already present in the resource.
func (v Var) EagerLoading(s *Spec) Var {
	s.Before(func(t *T) { _ = v.Get(t) })
	return v
}

// Append will append a value[T] to a current value of Var[[]T].
// Append only possible if the value type of Var is a slice type of T.
func Append(t *T, v Var, x ...interface{}) {
	rv := reflect.ValueOf(v.Get(t))
	var rx []reflect.Value
	for _, e := range x {
		rx = append(rx, reflect.ValueOf(e))
	}
	v.Set(t, reflect.Append(rv, rx...).Interface())
}
