package testcase

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
