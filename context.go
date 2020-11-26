package testcase

func newContext(parent *context) *context {
	return &context{
		hooks:     make([]hookBlock, 0),
		parent:    parent,
		vars:      newVariables(),
		immutable: false,
	}
}

type context struct {
	immutable     bool
	vars          *variables
	parent        *context
	hooks         []hookBlock
	parallel      bool
	sequential    bool
	skipBenchmark bool
	name          string
	description   string
	tags          []string
}

func (c *context) let(varName string, letBlock letBlock) {
	c.vars.defs[varName] = letBlock
}

func (c *context) isParallel() bool {
	var (
		isParallel   bool
		isSequential bool
	)

	for _, ctx := range c.allLinkListElement() {
		if ctx.parallel {
			isParallel = true
		}
		if ctx.sequential {
			isSequential = true
		}
	}

	return isParallel && !isSequential
}

func (c *context) allLinkListElement() []*context {
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
	for _, ctx := range c.allLinkListElement() {
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
