package testcase

import "testing"

// Steps provide you with the ability to create setup steps for your testing#T.Run based nested tests.
type Steps []func(*testing.T) func()

// AddWithTeardown create a new Steps object that should be stored in the current context with `:=`
// the function it receives should return a func() that will be used during `Setup` teardown.
func (s Steps) AddWithTeardown(step func(*testing.T) func()) Steps {
	return append(append(Steps{}, s...), step)
}

// Add create a new Steps object that should be stored in the current context with `:=`
func (s Steps) Add(step func(t *testing.T)) Steps {
	return s.AddWithTeardown(func(t *testing.T) func() {
		step(t)

		return func() {}
	})
}

// Setup execute all the hooks, and then return func that represent teardowns
// the returned function should be defered
func (s Steps) Setup(t *testing.T) func() {
	var teardowns []func()

	for _, steps := range s {
		teardowns = append(teardowns, steps(t))
	}

	return func() {
		for _, td := range teardowns {
			td()
		}
	}
}
