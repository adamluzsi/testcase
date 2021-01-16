package testcase

type visitor interface {
	addTestCase(testCase)
}

type visitable interface {
	acceptVisitor(visitor)
}

//--------------------------------------------------------------------------------------------------------------------//

type collector struct {
	testCases []testCase // ID => testCase
}

func (c *collector) addTestCase(tc testCase) {
	c.testCases = append(c.testCases, tc)
}

func (c *collector) getTestCases(v visitable) []testCase {
	v.acceptVisitor(c)
	return c.testCases
}
