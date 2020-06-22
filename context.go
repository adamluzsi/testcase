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
	vars        *variables
	parent      *context
	hooks       []hookBlock
	parallel    bool
	sequential  bool
	immutable   bool
	description string
}

func (c *context) let(varName string, letBlock func(*T) interface{}) {
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

const hookWarning = `you cannot create spec hooks after you used describe/when/and/then,
unless you create a new context with the previously mentioned calls`

func (c *context) addHook(h hookBlock) {
	if c.immutable {
		panic(hookWarning)
	}

	c.hooks = append(c.hooks, h)
}
