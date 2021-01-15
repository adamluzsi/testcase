package testcase

type visitor interface {
	addTestCase(id string, testCase func())
}

type visitable interface {
	acceptVisitor(visitor)
}

//--------------------------------------------------------------------------------------------------------------------//

type collector struct {
	testCases map[string]func() // ID => start test case
}

func (c *collector) getTestCases() map[string]func() {
	if c.testCases == nil {
		c.testCases = make(map[string]func())
	}
	return c.testCases
}

func (c *collector) addTestCase(id string, testCase func()) {
	c.getTestCases()[id] = testCase
}

func (c *collector) run(o orderer, v visitable) {
	v.acceptVisitor(c)

	names := make([]string, 0, len(c.getTestCases()))
	for name, _ := range c.getTestCases() {
		names = append(names, name)
	}
	o.Order(names)

	for _, name := range names {
		if start, ok := c.getTestCases()[name]; ok {
			start()
		}
	}
}
