package testcase

// TODO: update Ts to (type T) when Go2 released

// Var is a test helper structure, that allows easy way to access test runtime variables.
// In the future it will be updated to use Go2 type parameters.
type Var struct /* T */ {
	// Name is the test context variable name from where the cached value can be accessed later on.
	// Name is Mandatory when you create a variable, else the empty string will be used as the variable name.
	Name string
	// Init is an optional constructor definition that will be used when Var is bonded to a *Spec without constructor function passed to the Let function.
	// The goal of this field to initialize a variable that can be reused across different testing suites by bounding the Var to a given testing suite.
	//
	// Please use #Get if you wish to access a test runtime across cached variable value.
	// The value returned by this is not subject to any #Before and #Around hook that might mutate the variable value during the test runtime.
	// Init function doesn't cache the value in the test runtime context but literally just meant to initialize a value for the Var in a given test case.
	// Please use it with caution.
	Init letBlock /*[T]*/
}

// Get returns the current cached value of the given Variable
// When Go2 released, it will replace type casting
func (v Var) Get(t *T) (T interface{}) {
	return t.I(v.Name).(interface{}) // cast to T
}

// Set sets a value to a given variable during test runtime
func (v Var) Set(t *T, T interface{}) {
	t.Let(v.Name, T)
}

// EagerLoading allows the variable to be loaded before the action and assertion block is reached.
// This can be useful when you want to have a variable that cause side effect on your system.
// Like it should be present in some sort of attached resource/storage.
//
// For example you may persist the value in a storage as part of the initialization block,
// and then when the test/then block is reached, the entity is already present in the resource.
func (v Var) EagerLoading(s *Spec) {
	s.Before(func(t *T) { _ = v.Get(t) })
}

// Let allow you to set the variable value to a given context
func (v Var) Let(s *Spec, blk letBlock) {
	if blk == nil && v.Init != nil {
		s.Let(v.Name, v.Init)
		return
	}

	s.Let(v.Name, blk)
}

// LetValue set the value of the variable to a given block
func (v Var) LetValue(s *Spec, T interface{}) {
	s.LetValue(v.Name, T)
}
