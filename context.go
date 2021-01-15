package testcase

func newContext(parent *context) *context {
	c := &context{
		hooks:     make([]hookBlock, 0),
		parent:    parent,
		vars:      newVariables(),
		immutable: false,
	}
	if parent != nil {
		parent.children = append(parent.children, c)
	}
	return c
}

type context struct {
	parent   *context
	children []*context

	immutable     bool
	vars          *variables
	hooks         []hookBlock
	parallel      bool
	sequential    bool
	skipBenchmark bool
	retry         *Retry
	group         string
	description   string
	tags          []string
	tests         []test
}

func (c *context) let(varName string, letBlock letBlock) {
	c.vars.defs[varName] = letBlock
}

func (c *context) isParallel() bool {
	var (
		isParallel   bool
		isSequential bool
	)

	for _, ctx := range c.all() {
		if ctx.parallel {
			isParallel = true
		}
		if ctx.sequential {
			isSequential = true
		}
	}

	return isParallel && !isSequential
}

// visits context chain in a reverse order
// from children to parent direction
func (c *context) all() []*context {
	var (
		contexts []*context
		current  *context
	)

	current = c

	for {
		contexts = append([]*context{current}, contexts...)

		if current.parent != nil {
			current = current.parent
			continue
		}

		break
	}

	return contexts
}

func (c *context) getTagSet() map[string]struct{} {
	tagsSet := make(map[string]struct{})
	for _, ctx := range c.all() {
		for _, tag := range ctx.tags {
			tagsSet[tag] = struct{}{}
		}
	}
	return tagsSet
}

const hookWarning = `you cannot create spec hooks after you used describe/when/and/then,
unless you create a new context with the previously mentioned calls`

func (c *context) addHook(h hookBlock) {
	if c.immutable {
		panic(hookWarning)
	}

	c.hooks = append(c.hooks, h)
}

type test struct {
	id  string
	blk func()
}

func (c *context) addTest(id string, blk func()) {
	blk()
	//c.tests = append(c.tests, test{id: id, blk: blk})
}

func (c *context) acceptVisitor(v visitor) {
	for _, child := range c.children {
		child.acceptVisitor(v)
	}

	for _, test := range c.tests {
		v.addTestCase(test.id, test.blk)
	}
}
