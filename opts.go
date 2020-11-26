package testcase

func SkipBenchmark() testOption {
	return setupFunc(func(c *context) {
		c.skipBenchmark = true
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type testOption interface {
	setup(*context)
}

type setupFunc func(*context)

func (fn setupFunc) setup(c *context) { fn(c) }
