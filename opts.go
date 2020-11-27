package testcase

import "time"

// Flaky will mark the context/test as unstable.
// Whenever possible, try to fix flaky tests.
// Flaky test execution is tolerant towards failing assertion
// and they will be attempted to be re-ran in case of failure.
// Wait Timeout for a successful flaky test must be provided.
func Flaky(timeout time.Duration) ContextOption {
	return contextOptionFunc(func(c *context) {
		c.flaky = &flakyFlag{WaitTimeout: timeout}
	})
}

func SkipBenchmark() ContextOption {
	return contextOptionFunc(func(c *context) {
		c.skipBenchmark = true
	})
}

func Name(name string) ContextOption {
	return contextOptionFunc(func(c *context) {
		c.name = name
	})
}

func parallel() ContextOption {
	return contextOptionFunc(func(c *context) {
		c.parallel = true
	})
}

func sequential() ContextOption {
	return contextOptionFunc(func(c *context) {
		c.sequential = true
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ContextOption interface {
	setup(*context)
}

type contextOptionFunc func(*context)

func (fn contextOptionFunc) setup(c *context) { fn(c) }
