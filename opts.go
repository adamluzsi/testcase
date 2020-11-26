package testcase

func SkipBenchmark() option {
	return setupFunc(func(c *context) {
		c.skipBenchmark = true
	})
}

func Name(name string) option {
	return setupFunc(func(c *context) {
		c.name = name
	})
}

func parallel() option {
	return setupFunc(func(c *context) {
		c.parallel = true
	})
}

func sequential() option {
	return setupFunc(func(c *context) {
		c.sequential = true
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type option interface {
	setup(*context)
}

type setupFunc func(*context)

func (fn setupFunc) setup(c *context) { fn(c) }
