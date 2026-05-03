# testcase Spec user guide

## Spec Definition Hard Rules

This document defines the hard rules that must be followed when defining correct `testcase` specs. These rules are derived from the `testcase` framework's philosophy and best practices.

### Test Function Naming

#### Rule 1: The test function name MUST be the testing subject's name

```go
// ✅ Correct
func TestMyType(t *testing.T) { ... }
func TestParseInput(t *testing.T) { ... }
func TestRepository(t *testing.T) { ... }

// ❌ Incorrect
func TestSomething(t *testing.T) { ... }
func TestMainLogic(t *testing.T) { ... }
```

#### Rule 2: Test functions MUST follow the `Test<Subject>` convention

- Use PascalCase for the subject name
- Do not prefix with "Test" more than once
- Keep it concise and descriptive

---

### Spec Declaration

#### Rule 3: Create spec immediately in test function body

```go
func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)
    
    // ... spec definition
}
```

#### Rule 4: Do not defer spec creation

```go
// ❌ Incorrect
func TestMyType(t *testing.T) {
    var s *testcase.Spec
    t.Run("setup", func(t *testing.T) {
        s = testcase.NewSpec(t)
    })
}

// ✅ Correct
func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)
}
```

---

### Subject Definition

#### Rule 5: Define the System Under Test (SUT) at spec root level

```go
func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)
    
    subject := let.Var(s, func(t *testcase.T) MyType {
        return MyType{}
    })
}
```

#### Rule 6: Subject MUST be accessible via getter function or `.Get(t)`

```go
// Using Var
subject := let.Var(s, func(t *testcase.T) MyType { ... })
// Access: subject.Get(t)

// Or helper function
func subject(t *testcase.T) MyType { ... }
```

---

### Describe Blocks

#### Rule 7: A `Spec#Describe` MUST have a dedicated `act` defined as a function

The `act` function takes `*testcase.T` only as its input and executes the operation that the current `Spec#Describe` block is about.

```go
s.Describe("#MyFunc", func(s *testcase.Spec) {
    var (
        input = let.Var(s, func(t *testcase.T) string {
            return t.Random.String()
        })
    )
    
    // ✅ Correct: act defined at top of Describe block
    act := func(t *testcase.T) string {
        return subject.Get(t).MyFunc(input.Get(t))
    }
    
    s.Then("returns input unchanged", func(t *testcase.T) {
        assert.Equal(t, act(t), input.Get(t))
    })
})
```

#### Rule 8: Each `Describe` block MUST have exactly one ACT

```go
// ❌ Incorrect - multiple different acts in same Describe
s.Describe("#MyFunc", func(s *testcase.Spec) {
    s.Then("test 1", func(t *testcase.T) {
        result := subject.Get(t).MyFunc(input.Get(t))
    })
    
    s.Then("test 2", func(t *testcase.T) {
        err := subject.Get(t).Save(ctx.Get(t), entity.Get(t)) // Different method!
    })
})

// ✅ Correct - one ACT per Describe block
s.Describe("#MyFunc", func(s *testcase.Spec) {
    act := func(t *testcase.T) string {
        return subject.Get(t).MyFunc(input.Get(t))
    }
    
    s.Then("test 1", func(t *testcase.T) {
        _ = act(t)
    })
})

s.Describe("#Save", func(s *testcase.Spec) {
    act := func(t *testcase.T) error {
        return subject.Get(t).Save(ctx.Get(t), entity.Get(t))
    }
    
    s.Then("test 2", func(t *testcase.T) {
        _ = act(t)
    })
})
```

#### Rule 9: Describe blocks MUST group tests by method or feature

```go
s.Describe("#MyFunc", func(s *testcase.Spec) { ... })
s.Describe("#Save", func(s *testcase.Spec) { ... })
s.Describe("validation", func(s *testcase.Spec) { ... })
```

---

### Variables and Inputs

#### Rule 10: Test variables MUST be defined using `let.Var` or `let.Let`

```go
var (
    input = let.Var(s, func(t *testcase.T) string {
        return t.Random.String()
    })
    
    expected = let.Var(s, func(t *testcase.T) string {
        return input.Get(t)
    })
)
```

#### Rule 11: Variables MUST be accessed via `.Get(t)` in test functions

```go
// ✅ Correct
s.Then("returns expected value", func(t *testcase.T) {
    assert.Equal(t, act(t), input.Get(t))
})

// ❌ Incorrect - capturing variable outside test
var capturedInput string
s.Before(func(t *testcase.T) {
    capturedInput = input.Get(t) // Wrong!
})
```

#### Rule 12: Use `t.Random` for property testing to enable reproducible failures

```go
input := t.Random.String()           // Random string
id := t.Random.UUID()                // Random UUID  
num := t.Random.IntBetween(1, 100)   // Random int in range
data := t.Random.Bytes(32)           // Random bytes
```

---

### The ACT Pattern

#### Rule 13: `act` MUST be defined at the top of each `Describe` block

```go
s.Describe("#MyFunc", func(s *testcase.Spec) {
    // ✅ Correct - act at top
    act := func(t *testcase.T) string {
        return subject.Get(t).MyFunc(input.Get(t))
    }
    
    var input = let.Var(s, ...)
    
    s.Then("...", func(t *testcase.T) { ... })
})

// ❌ Incorrect - act defined after tests
s.Describe("#MyFunc", func(s *testcase.Spec) {
    s.Then("...", func(t *testcase.T) { ... }) // Can't use act yet!
    
    act := func(t *testcase.T) string { ... }
})
```

#### Rule 14: `act` MUST be immutable - no side effects in test functions

```go
// ✅ Correct - act is pure function
act := func(t *testcase.T) string {
    return subject.Get(t).MyFunc(input.Get(t))
}

s.Then("test", func(t *testcase.T) {
    result := act(t) // Just call it
    assert.Equal(t, result, expected.Get(t))
})

// ❌ Incorrect - modifying state in test
act := func(t *testcase.T) string {
    globalCounter++ // Side effect!
    return subject.Get(t).MyFunc(input.Get(t))
}
```

#### Rule 15: `act` MUST NOT vary within tests of the same Describe block

```go
// ❌ Incorrect - varying ACT within tests
s.Then("test 1", func(t *testcase.T) {
    result := subject.Get(t).MyFunc(input.Get(t))
})

s.Then("test 2", func(t *testcase.T) {
    result := subject.Get(t).MyFunc(otherInput) // Different input!
})

// ✅ Correct - ACT is consistent, context varies via When/And
act := func(t *testcase.T) string {
    return subject.Get(t).MyFunc(input.Get(t))
}

s.Then("test 1", func(t *testcase.T) {
    assert.Equal(t, act(t), input.Get(t))
})

s.When("input is different", func(s *testcase.Spec) {
    input.LetValue(s, "different")
    
    s.Then("test 2", func(t *testcase.T) {
        assert.Equal(t, act(t), input.Get(t)) // Same ACT, different context
    })
})
```

---

### Then Clauses (Happy Path)

#### Rule 16: `Spec#Then` / `Spec#Test` should use the current scope's `act`

```go
s.Describe("#MyFunc", func(s *testcase.Spec) {
    act := func(t *testcase.T) string {
        return subject.Get(t).MyFunc(input.Get(t))
    }
    
    // ✅ Correct - uses act from Describe scope
    s.Then("returns input unchanged", func(t *testcase.T) {
        assert.Equal(t, act(t), input.Get(t))
    })
})
```

#### Rule 17: Happy path tests MUST NOT be wrapped in `When` blocks

```go
// ❌ Incorrect - wrapping happy path in When
s.Describe("#MyFunc", func(s *testcase.Spec) {
    act := func(t *testcase.T) string { ... }
    
    s.When("happy path", func(s *testcase.Spec) {
        s.Then("returns expected value", func(t *testcase.T) {
            assert.Equal(t, act(t), input.Get(t))
        })
    })
})

// ✅ Correct - happy path at Describe level
s.Describe("#MyFunc", func(s *testcase.Spec) {
    act := func(t *testcase.T) string { ... }
    
    s.Then("returns expected value", func(t *testcase.T) {
        assert.Equal(t, act(t), input.Get(t))
    })
})
```

#### Rule 18: `Then` descriptions MUST express the expected behavior

```go
s.Then("returns the input value unchanged", func(t *testcase.T) { ... })
s.Then("creates a new entity with generated ID", func(t *testcase.T) { ... })
s.Then("persists the entity to storage", func(t *testcase.T) { ... })

// ❌ Avoid vague descriptions
s.Then("works correctly", func(t *testcase.T) { ... })
s.Then("does something", func(t *testcase.T) { ... })
```

#### Rule 19: Happy path tests MUST come first in Describe blocks

```go
s.Describe("#MyFunc", func(s *testcase.Spec) {
    act := func(t *testcase.T) string { ... }
    
    // ✅ Correct - happy path first
    s.Then("returns input unchanged", func(t *testcase.T) { ... })
    
    // Rainy paths after
    s.When("input is empty", func(s *testcase.Spec) { ... })
})
```

---

### When/And Contexts (Rainy Paths)

#### Rule 20: A `Spec#Context` or `#When`/`#And` MUST have a test arrangement

The context block must modify variables or set up hooks to create the alternative condition.

```go
// ✅ Correct - When has arrangement
s.When("context is canceled", func(s *testcase.Spec) {
    spec.ctx().Let(s, func(t *testcase.T) context.Context {
        c, cancel := context.WithCancel(context.Background())
        cancel() // Arrange: cancel the context
        return c
    })
    
    s.Then("returns context.Canceled error", func(t *testcase.T) { ... })
})

// ❌ Incorrect - When without arrangement
s.When("some condition", func(s *testcase.Spec) {
    s.Then("expectation", func(t *testcase.T) { ... }) // No setup!
})
```

#### Rule 21: `When` blocks MUST start with variable modification or hook definition

```go
s.When("input is empty", func(s *testcase.Spec) {
    // ✅ Correct - starts with arrangement
    input.LetValue(s, "")
    
    s.Then("returns error", func(t *testcase.T) { ... })
})

s.When("during transaction", func(s *testcase.Spec) {
    // ✅ Correct - starts with hook
    s.Before(func(t *testcase.T) {
        tx := beginTransaction(t)
        spec.ctx().Set(t, tx)
    })
    
    s.Then("events are buffered", func(t *testcase.T) { ... })
})
```

#### Rule 22: `And` blocks MUST add additional conditions within a `When` context

```go
s.When("input is empty", func(s *testcase.Spec) {
    input.LetValue(s, "")
    
    s.And("validation is enabled", func(s *testcase.Spec) {
        subject.Let(s, func(t *testcase.T) MyType {
            sub := subject.Super(t)
            sub.Validate = true
            return sub
        })
        
        s.Then("returns validation error", func(t *testcase.T) { ... })
    })
})
```

#### Rule 23: `When`/`And` nesting MUST remain shallow (max 2-3 levels)

```go
// ❌ Incorrect - too deep!
s.When("condition A", func(s *testcase.Spec) {
    s.And("and condition B", func(s *testcase.Spec) {
        s.And("but also condition C", func(s *testcase.Spec) {
            s.And("finally condition D", func(s *testcase.Spec) {
                s.Then("expectation", func(t *testcase.T) { ... })
            })
        })
    })
})

// ✅ Correct - flat structure
s.Test("with conditions A, B, C, and D", func(t *testcase.T) {
    // Setup all conditions inline
    // Assert expectation
})

s.When("condition A fails", func(s *testcase.Spec) {
    s.Then("returns error", func(t *testcase.T) { ... })
})
```

---

### Before Hooks

#### Rule 24: `s.Before()` hooks MUST be defined before test assertions

```go
s.When("during transaction", func(s *testcase.Spec) {
    // ✅ Correct - Before defined before Then
    s.Before(func(t *testcase.T) {
        tx, err := spec.memoryGet(t).BeginTx(spec.ctxGet(t))
        assert.Must(t).NoError(err)
        spec.ctx().Set(t, tx)
    })
    
    s.Then("events are buffered until commit", func(t *testcase.T) { ... })
})
```

#### Rule 25: Before hooks MUST be idempotent and reversible

```go
// ✅ Correct - hook sets up clean state
s.Before(func(t *testcase.T) {
    cleanup := setupResources(t)
    t.Cleanup(cleanup) // Ensure reversal
})

// ❌ Incorrect - hook accumulates state
var counter int
s.Before(func(t *testcase.T) {
    counter++ // State leak between tests!
})
```

---

### Nesting Structure

#### Rule 26: Keep nesting flat - happy path is the default context

```go
// ✅ Correct - flat structure with happy path first
func TestMerge(t *testing.T) {
    s := testcase.NewSpec(t)
    
    act := func(t *testcase.T) error { ... }
    
    // Happy path - no When() wrapper
    s.Test("returns nil when no errors", func(t *testcase.T) { ... })
    
    // Rainy paths branch from base
    s.When("an error is supplied", func(s *testcase.Spec) { ... })
}

// ❌ Incorrect - wrapping happy path in When
func TestMerge(t *testing.T) {
    s := testcase.NewSpec(t)
    
    act := func(t *testcase.T) error { ... }
    
    s.When("happy path", func(s *testcase.Spec) { // Wrong!
        s.Test("returns nil", func(t *testcase.T) { ... })
    })
}
```

#### Rule 27: Use `s.Test()` for independent scenarios, `s.Then()` within contexts

```go
// ✅ Correct - Test() at Describe level for independent tests
s.Describe("#MyFunc", func(s *testcase.Spec) {
    act := func(t *testcase.T) string { ... }
    
    s.Test("happy path works", func(t *testcase.T) { ... })
    s.Test("another independent scenario", func(t *testcase.T) { ... })
})

// ✅ Correct - Then() within When/And contexts
s.When("condition is met", func(s *testcase.Spec) {
    s.Then("expectation holds", func(t *testcase.T) { ... })
})
```

---

### Test Independence

#### Rule 28: Tests MUST NOT depend on execution order

```go
// ❌ Incorrect - test depends on previous test's side effect
var counter int

func TestSomething(t *testing.T) {
    s := testcase.NewSpec(t)
    
    s.Test("increments counter", func(t *testcase.T) {
        counter++ // Shared mutable state!
        assert.Equal(t, counter, 1)
    })
    
    s.Test("counter is still incremented", func(t *testcase.T) {
        assert.Equal(t, counter, 2) // Depends on order!
    })
}

// ✅ Correct - each test is independent
func TestSomething(t *testing.T) {
    s := testcase.NewSpec(t)
    
    var counter = let.Var(s, func(t *testcase.T) int {
        return t.Random.Int()
    })
    
    s.Test("handles counter value", func(t *testcase.T) {
        c := counter.Get(t)
        // Test with this specific value
    })
}
```

#### Rule 29: Tests MUST NOT use numbered naming conventions to imply order

```go
// ❌ Incorrect - implies execution order
func Test01Initialize(t *testing.T) { ... }
func Test02Process(t *testing.T) { ... }
func Test03Cleanup(t *testing.T) { ... }

// ✅ Correct - descriptive names, no implied order
func TestInitializeCreatesValidState(t *testing.T) { ... }
func TestProcessWithValidInputSucceeds(t *testing.T) { ... }
func TestCleanupReleasesResources(t *testing.T) { ... }
```

#### Rule 30: Each test MUST set up its own required state

```go
// ❌ Incorrect - relies on parent spec state
s.Describe("#MyFunc", func(s *testcase.Spec) {
    input := let.Var(s, func(t *testcase.T) string { return "default" })
    
    s.When("input is empty", func(s *testcase.Spec) {
        // Forgot to override input!
        
        s.Then("returns error", func(t *testcase.T) { ... })
    })
})

// ✅ Correct - explicit state setup in each context
s.Describe("#MyFunc", func(s *testcase.Spec) {
    input := let.Var(s, func(t *testcase.T) string { return "default" })
    
    s.When("input is empty", func(s *testcase.Spec) {
        input.LetValue(s, "") // Explicitly override
        
        s.Then("returns error", func(t *testcase.T) { ... })
    })
})
```

---

### Summary Checklist

Before submitting a spec, verify:

- [ ] Test function name matches the testing subject (`Test<Subject>`)
- [ ] Spec is created immediately with `testcase.NewSpec(t)`
- [ ] System Under Test (subject) is defined at spec root level
- [ ] Each `Describe` block has exactly one `act` function defined at the top
- [ ] `act` function takes only `*testcase.T` as input
- [ ] Happy path tests use `s.Then()` or `s.Test()` directly (not wrapped in `When`)
- [ ] Happy path tests come first in each `Describe` block
- [ ] Each `When`/`And` context has explicit arrangement (variable modification or hook)
- [ ] Nesting remains flat (max 2-3 levels deep)
- [ ] Tests use `.Get(t)` to access variables inside test functions
- [ ] No shared mutable state between tests
- [ ] No numbered test names implying execution order
- [ ] Each test is independent and can run in any order

## testcase Spec Schema

A formal specification describing how behavioral tests should be structured using the `testcase` framework. This document defines conventions, patterns, and rules that ensure consistency across the codebase.

---

### Table of Contents

1. [Philosophy](#philosophy)
2. [File Structure](#file-structure)
3. [Test Function Naming](#test-function-naming)
4. [Spec Declaration](#spec-declaration)
5. [Subject Definition](#subject-definition)
6. [Describe Blocks](#describe-blocks)
7. [Variables and Inputs](#variables-and-inputs)
8. [The ACT Pattern](#the-act-pattern)
9. [Then Clauses (Happy Path)](#then-clauses-happy-path)
10. [When/And Contexts (Rainy Paths)](#whenand-contexts-rainy-paths)
11. [Before Hooks](#before-hooks)
12. [Complete Example](#complete-example)
13. [Common Patterns](#common-patterns)
14. [Anti-Patterns to Avoid](#anti-patterns-to-avoid)
15. [Random Execution Order](#random-execution-order)

---

### Philosophy

#### Black-Box Testing First

Always test from the consumer's perspective. You are the very first user of your own API.
Test files should use the `_test` package suffix for black-box testing.

```go
// Recommended - black box testing
package mypkg_test

import (
    "testing"
    
    "mypkg"
    "go.llib.dev/testcase"
)

func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)
    // Test the public API, not internal implementation
}
```

---

#### Happy Path First, Keep It Flat

The most simplistic happy path should be at the top level of any `Describe` block. Rainy paths branch from this base using `When`/`And`:

```go
s.Describe("#Method", func(s *testcase.Spec) {
    act := func(t *testcase.T) (...) {...}

    // Happy path - default context, no When() wrapper
    s.Then("returns expected value", func(t *testcase.T) {
        // ...
    })
    
    // Rainy paths - variations from base
    s.When("condition is invalid", func(s *testcase.Spec) {
        s.Then("returns error", func(t *testcase.T) {
            // ...
        })
    })
})
```

---

### Test Function Naming

#### Rule: The name MUST be the testing subject's name

- If testing a Type (e.g., `MyType`), use `TestMyType`
- If testing a function (e.g., `MyFunc`), use `TestMyFunc`
- For sub-testing contexts, express with `testcase.Spec` context blocks (`Describe`, `When`, etc.)

```go
// Type under test
func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)
    // ...
}

// Function under test  
func TestParseInput(t *testing.T) {
    s := testcase.NewSpec(t)
    // ...
}

// Method under test - still use the type name, describe method in Describe block
func TestRepository(t *testing.T) {
    s := testcase.NewSpec(t)
    
    s.Describe("#Create", func(s *testcase.Spec) {
        // Create-specific tests
    })
    
    s.Describe("#FindByID", func(s *testcase.Spec) {
        // FindByID-specific tests
    })
}
```

---

### Spec Declaration

Every test function starts by creating a spec from `*testing.T`:

```go
func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)
    
    // Define variables, describe blocks, etc.
}
```

---

### Subject Definition

The **subject** is the system under test (SUT). Define it at the spec level using `let.Var`:

```go
subject := let.Var(s, func(t *testcase.T) mypkg.MyType {
    return mypkg.MyType{}
})

// Access in tests with subject.Get(t)
s.Then("subject is initialized", func(t *testcase.T) {
    assert.NotNil(t, subject.Get(t))
})
```

#### Subject Variations

Override the subject in specific contexts when needed:

```go
s.When("ToUpper option is set", func(s *testcase.Spec) {
    subject.Let(s, func(t *testcase.T) mypkg.MyType {
        sub := subject.Super(t)  // Get parent value
        sub.ToUpper = true       // Modify for this context
        return sub
    })
    
    s.Then("output is uppercased", func(t *testcase.T) {
        // ...
    })
})
```

---

### Describe Blocks

#### Purpose

`Describe` blocks group related tests around a specific method or feature:

```go
s.Describe("#MyFunc", func(s *testcase.Spec) {
    // Tests for MyFunc go here
})
```

#### Rules for Describe Blocks

1. **Must have a dedicated ACT** defined at the top of the block
2. The ACT should express what the describe block is meant to test
3. ACT might have input arguments if the method requires them

```go
s.Describe("#MyFunc", func(s *testcase.Spec) {
    var (
        input = let.Var(s, func(t *testcase.T) string {
            return t.Random.String()
        })
    )
    
    // Dedicated ACT for this Describe block
    act := func(t *testcase.T) string {
        return subject.Get(t).MyFunc(input.Get(t))
    }
    
    s.Then("input returned as is", func(t *testcase.T) {
        assert.Equal(t, act(t), input.Get(t))
    })
})
```

---

### Variables and Inputs

#### Using `let.Var` for Test Variables

Define test variables with `let.Var`:

```go
var (
    input = let.Var(s, func(t *testcase.T) string {
        return t.Random.String()  // Pseudo-deterministic random
    })
    
    expected = let.Var(s, func(t *testcase.T) string {
        return input.Get(t)  // Derived from other variables
    })
)
```

#### Using `t.Random` for Property Testing

`t.Random` provides pseudo-deterministic random data:

```go
input := t.Random.String()           // Random string
id := t.Random.UUID()                // Random UUID  
num := t.Random.IntBetween(1, 100)   // Random int in range
data := t.Random.Bytes(32)           // Random bytes
```

**Benefit:** If a test fails due to an unhandled random input, you get the `TESTCASE_SEED` that can recreate the failing scenario 1:1.

#### Variable Access Patterns

| Method | Purpose | Example |
|--------|---------|---------|
| `.Get(t)` | Get current value | `input.Get(t)` |
| `.Set(t, v)` | Set value in this spec | `input.Set(t, "value")` |
| `.Let(s, fn)` | Override in nested spec | `input.Let(s, func(t) { ... })` |
| `.LetValue(s, v)` | Override with constant | `input.LetValue(s, nil)` |
| `.Super(t)` | Get parent value | `subject.Super(t)` |

---

### The ACT Pattern

#### Definition

The **ACT** is an immutable testing function that represents the action being tested:

```go
act := func(t *testcase.T) string {
    return subject.Get(t).MyFunc(input.Get(t))
}
```

#### Rules for ACT

1. **Must be defined at the top of each `Describe` block**
2. **Must be immutable** - no side effects in test functions themselves
3. **Centralizes arrangement** - forces you to arrange inputs and context upfront
4. **Reduces mental model complexity** - tests focus on context building, not ACT construction

#### Why the ACT Pattern Matters

```go
// ❌ Anti-pattern: ACT varies within tests
s.Then("test 1", func(t *testcase.T) {
    result := subject.Get(t).MyFunc(input.Get(t))
    // ...
})

s.Then("test 2", func(t *testcase.T) {
    result := subject.Get(t).MyFunc(otherInput)  // Different input!
    // ...
})

// ✅ Correct: ACT is consistent, context varies
act := func(t *testcase.T) string {
    return subject.Get(t).MyFunc(input.Get(t))
}

s.Then("test 1", func(t *testcase.T) {
    result := act(t)
    // ...
})

s.When("input is different", func(s *testcase.Spec) {
    input.LetValue(s, "different")
    
    s.Then("test 2", func(t *testcase.T) {
        result := act(t)  // Same ACT, different context
        // ...
    })
})
```

---

### Then Clauses (Happy Path)

#### Placement

`Then` clauses expressing the **most simplistic happy path** should be at the top level of a `Describe` block:

```go
s.Describe("#MyFunc", func(s *testcase.Spec) {
    // ... variable and ACT definitions ...
    
    // Happy path - no When() wrapper needed
    s.Then("input returned as is", func(t *testcase.T) {
        assert.Equal(t, act(t), input.Get(t))
    })
})
```

#### Naming Convention

`Then` descriptions should express the expected behavior:

```go
s.Then("returns the input value unchanged", func(t *testcase.T) { ... })
s.Then("creates a new entity with generated ID", func(t *testcase.T) { ... })
s.Then("persists the entity to storage", func(t *testcase.T) { ... })
```

---

### When/And Contexts (Rainy Paths)

#### Purpose

`When` and `And` blocks describe alternative conditions or states that differ from the happy path:

```go
s.When("context is canceled", func(s *testcase.Spec) {
    // Arrange context for this scenario
    
    s.Then("returns context.Canceled error", func(t *testcase.T) {
        // Assert expected behavior
    })
})
```

#### Rules for When/And

1. **Must start with arrangement** - modify variables or set up hooks
2. **Can nest `And` blocks** for additional conditions (but keep it shallow!)
3. **Inherits parent context** - all parent variables and hooks are available

#### Variable Modification in Contexts

```go
s.When("MyType#ToUpper option is set", func(s *testcase.Spec) {
    subject.Let(s, func(t *testcase.T) mypkg.MyType {
        sub := subject.Super(t)  // Get parent value
        sub.ToUpper = true       // Modify for this context
        return sub
    })

    s.Then("the result will be in upper format", func(t *testcase.T) {
        assert.Equal(t, act(t), strings.ToUpper(input.Get(t)))
    })
})
```

#### Using `And` for Additional Conditions

```go
s.When("input is empty", func(s *testcase.Spec) {
    input.LetValue(s, "")

    s.And("validation is enabled", func(s *testcase.Spec) {
        subject.Let(s, func(t *testcase.T) mypkg.MyType {
            sub := subject.Super(t)
            sub.Validate = true
            return sub
        })

        s.Then("returns validation error", func(t *testcase.T) {
            assert.ErrorIs(t, act(t), ErrValidation)
        })
    })
})
```

---

### Before Hooks

#### Purpose

`s.Before()` hooks run before each test in the context. Use them for:
- Setup that applies to multiple tests
- Cleanup registration via `t.Cleanup()`
- Debug logging on failure via `t.OnFail()`

#### Hook Execution Order

Hooks execute **sequentially from outermost to innermost**:

```go
s.Describe("#Outer", func(s *testcase.Spec) {
    s.Before(func(t *testcase.T) {
        t.Log("outer before")  // Runs first
    })
    
    s.When("inner context", func(s *testcase.Spec) {
        s.Before(func(t *testcase.T) {
            t.Log("inner before")  // Runs second
        })
        
        s.Then("test runs", func(t *testcase.T) {
            t.Log("test")  // Runs last
        })
    })
})

// Output order: "outer before" -> "inner before" -> "test"
```

#### Common Before Hook Patterns

**Debug logging on failure:**

```go
s.Before(func(t *testcase.T) {
    args := slicekit.Clone(request.Get(t).Args)
    t.OnFail(func() {
        t.Log("args:", args)
        t.Log("code:", response.Get(t).Code)
        t.Log("\nout:\n", response.Get(t).Out.String())
        t.Log("\nerr:\n", response.Get(t).Err.String())
    })
})
```

**Transaction setup:**

```go
s.When("during transaction", func(s *testcase.Spec) {
    s.Before(func(t *testcase.T) {
        tx, err := subject.Get(t).BeginTx(ctx)
        assert.Must(t).NoError(err)
        ctxVar.Set(t, tx)  // Override context with transaction
    })

    s.Then("changes are buffered", func(t *testcase.T) {
        // ...
    })
})
```

---

### Complete Example

```go
package mypkg_test

import (
    "strings"
    "testing"

    "mypkg"
    
    "go.llib.dev/testcase"
    "go.llib.dev/testcase/assert"
    "go.llib.dev/testcase/let"
)

// TestMyType tests the MyType system under test.
func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)

    // Subject: the system under test
    subject := let.Var(s, func(t *testcase.T) mypkg.MyType {
        return mypkg.MyType{}
    })

    s.Describe("#MyFunc", func(s *testcase.Spec) {
        var (
            // Input variable with pseudo-random data
            input = let.Var(s, func(t *testcase.T) string {
                return t.Random.String()
            })
        )

        // ACT: the immutable testing action for this Describe block
        act := func(t *testcase.T) string {
            return subject.Get(t).MyFunc(input.Get(t))
        }

        // Happy path - default behavior
        s.Then("input returned as is", func(t *testcase.T) {
            assert.Equal(t, act(t), input.Get(t))
        })

        // Rainy path: when ToUpper option is set
        s.When("MyType#ToUpper option is set", func(s *testcase.Spec) {
            subject.Let(s, func(t *testcase.T) mypkg.MyType {
                sub := subject.Super(t)
                sub.ToUpper = true
                return sub
            })

            s.Then("the result will be in upper format", func(t *testcase.T) {
                assert.Equal(t, act(t), strings.ToUpper(input.Get(t)))
            })
        })

        // Rainy path: when input is empty
        s.When("input is empty string", func(s *testcase.Spec) {
            input.LetValue(s, "")

            s.Then("returns empty string", func(t *testcase.T) {
                assert.Equal(t, act(t), "")
            })
        })
    })
}
```

#### System Under Test

```go
package mypkg

import "strings"

type MyType struct {
    ToUpper bool
}

func (mt MyType) MyFunc(v string) string {
    if mt.ToUpper {
        return strings.ToUpper(v)
    }
    return v
}
```

---

### Common Patterns

#### Pattern 1: Testing Context-Aware Functions

```go
s.Describe("#Process", func(s *testcase.Spec) {
    var (
        ctx = let.Var(s, func(t *testcase.T) context.Context {
            return context.Background()
        })
        
        input = let.Var(s, func(t *testcase.T) string {
            return t.Random.String()
        })
    )

    act := func(t *testcase.T) error {
        return subject.Get(t).Process(ctx.Get(t), input.Get(t))
    }

    // Happy path
    s.Then("processes successfully", func(t *testcase.T) {
        assert.NoError(t, act(t))
    })

    // Context canceled scenario
    s.When("context is canceled", func(s *testcase.Spec) {
        ctx.Let(s, func(t *testcase.T) context.Context {
            c, cancel := context.WithCancel(context.Background())
            cancel()  // Cancel immediately
            return c
        })

        s.Then("returns context.Canceled error", func(t *testcase.T) {
            assert.ErrorIs(t, act(t), context.Canceled)
        })
    })
})
```

#### Pattern 2: Testing with Transactions

```go
s.Describe("#Save", func(s *testcase.Spec) {
    var (
        entity = let.Var(s, func(t *testcase.T) *mypkg.Entity {
            return &mypkg.Entity{ID: t.Random.UUID()}
        })
        
        ctx = let.Var(s, func(t *testcase.T) context.Context {
            return context.Background()
        })
    )

    act := func(t *testcase.T) error {
        return subject.Get(t).Save(ctx.Get(t), entity.Get(t))
    }

    // Happy path: background context
    s.Then("entity is persisted", func(t *testcase.T) {
        assert.NoError(t, act(t))
        assert.True(t, subject.Get(t).Exists(ctx.Get(t), entity.Get(t).ID))
    })

    // Transaction scenario
    s.When("called within transaction", func(s *testcase.Spec) {
        s.Before(func(t *testcase.T) {
            tx, err := subject.Get(t).BeginTx(ctx.Get(t))
            assert.Must(t).NoError(err)
            ctx.Set(t, tx)  // Override with transaction context
            
            t.Cleanup(func() {
                _ = subject.Get(t).RollbackTx(tx)
            })
        })

        s.Then("entity is buffered until commit", func(t *testcase.T) {
            assert.NoError(t, act(t))
            assert.False(t, subject.Get(t).Exists(ctx.Super(t), entity.Get(t).ID))  // Not visible outside
            
            assert.NoError(t, subject.Get(t).CommitTx(ctx.Get(t)))
            assert.True(t, subject.Get(t).Exists(ctx.Super(t), entity.Get(t).ID))  // Now visible
        })
    })
})
```

#### Pattern 3: Testing Error Cases with Nested Conditions

```go
s.Describe("#Validate", func(s *testcase.Spec) {
    var (
        input = let.Var(s, func(t *testcase.T) string {
            return t.Random.String()
        })
    )

    act := func(t *testcase.T) error {
        return subject.Get(t).Validate(input.Get(t))
    }

    // Happy path
    s.Then("returns nil for valid input", func(t *testcase.T) {
        assert.NoError(t, act(t))
    })

    // Error cases
    s.When("input is empty", func(s *testcase.Spec) {
        input.LetValue(s, "")

        s.Then("returns ErrEmptyInput", func(t *testcase.T) {
            assert.ErrorIs(t, act(t), mypkg.ErrEmptyInput)
        })
    })

    s.When("input exceeds maximum length", func(s *testcase.Spec) {
        input.Let(s, func(t *testcase.T) string {
            return t.Random.StringN(1001)  // Max is 1000
        })

        s.Then("returns ErrInputTooLong", func(t *testcase.T) {
            assert.ErrorIs(t, act(t), mypkg.ErrInputTooLong)
        })
    })

    s.When("input contains invalid characters", func(s *testcase.Spec) {
        input.LetValue(s, "invalid@char#acters!")

        s.And("strict mode is enabled", func(s *testcase.Spec) {
            subject.Let(s, func(t *testcase.T) mypkg.Validator {
                v := subject.Super(t)
                v.Strict = true
                return v
            })

            s.Then("returns ErrInvalidCharacters", func(t *testcase.T) {
                assert.ErrorIs(t, act(t), mypkg.ErrInvalidCharacters)
            })
        })
    })
})
```

---

### Anti-Patterns to Avoid

#### ❌ Anti-Pattern: Wrapping Happy Path in `When`

```go
// Don't do this
s.When("happy path", func(s *testcase.Spec) {
    s.Then("returns expected value", func(t *testcase.T) {
        // ...
    })
})

// Do this instead
s.Then("returns expected value", func(t *testcase.T) {
    // ...
})
```

#### ❌ Anti-Pattern: Deep Nesting

```go
// Don't do this - too deep!
s.When("condition A", func(s *testcase.Spec) {
    s.And("and condition B", func(s *testcase.Spec) {
        s.And("but also condition C", func(s *testcase.Spec) {
            s.And("finally condition D", func(s *testcase.Spec) {
                s.Then("expectation", func(t *testcase.T) {
                    // ...
                })
            })
        })
    })
})

// Do this instead - flat structure
s.Test("with conditions A, B, C, and D", func(t *testcase.T) {
    // Setup all conditions
    // Assert expectation
})

s.When("condition A fails", func(s *testcase.Spec) {
    s.Then("returns error", func(t *testcase.T) {
        // ...
    })
})
```

#### ❌ Anti-Pattern: Varying ACT Within Tests

```go
// Don't do this
s.Then("test 1", func(t *testcase.T) {
    result := subject.Get(t).MyFunc(input.Get(t))
    assert.Equal(t, result, input.Get(t))
})

s.Then("test 2", func(t *testcase.T) {
    result := subject.Get(t).MyFunc(otherInput)  // Different!
    assert.Equal(t, result, otherInput)
})

// Do this instead
act := func(t *testcase.T) string {
    return subject.Get(t).MyFunc(input.Get(t))
}

s.Then("test 1", func(t *testcase.T) {
    assert.Equal(t, act(t), input.Get(t))
})

s.When("input is different", func(s *testcase.Spec) {
    input.LetValue(s, otherInput)
    
    s.Then("test 2", func(t *testcase.T) {
        assert.Equal(t, act(t), input.Get(t))
    })
})
```

#### ❌ Anti-Pattern: Missing ACT Definition

```go
// Don't do this - no clear ACT
s.Describe("#MyFunc", func(s *testcase.Spec) {
    s.Then("test 1", func(t *testcase.T) {
        result := subject.Get(t).MyFunc(input.Get(t))
        // ...
    })
    
    s.Then("test 2", func(t *testcase.T) {
        err := subject.Get(t).Save(ctx.Get(t), entity.Get(t))  // Different method!
        // ...
    })
})

// Do this instead - one ACT per Describe block
s.Describe("#MyFunc", func(s *testcase.Spec) {
    act := func(t *testcase.T) string {
        return subject.Get(t).MyFunc(input.Get(t))
    }
    
    s.Then("test 1", func(t *testcase.T) {
        _ = act(t)
        // ...
    })
})

s.Describe("#Save", func(s *testcase.Spec) {
    act := func(t *testcase.T) error {
        return subject.Get(t).Save(ctx.Get(t), entity.Get(t))
    }
    
    s.Then("test 2", func(t *testcase.T) {
        _ = act(t)
        // ...
    })
})
```

---

### Random Execution Order

#### Why Random Order Matters

The `testcase` framework executes tests in a **random order by default**. This is intentional and critical for test quality:

- **Flaws hidden by ordering**: Tests that pass in one order may fail in another if they share mutable state or have implicit dependencies
- **Independence enforcement**: Random execution forces you to write truly independent tests with no assumptions about execution sequence
- **Early detection of coupling**: Tests that depend on each other will fail quickly and visibly when shuffled

#### The TESTCASE_SEED

Each spec has a `TESTCASE_SEED` that ensures:

1. **Reproducible failures**: When a test fails due to ordering issues, the seed is printed in the output
2. **Debug with same order**: Re-run with `-testcase.seed=<value>` to reproduce the exact failing sequence
3. **Confidence in fixes**: Verify your fix works by running with the same seed that caused the failure

```bash
## Run tests (random order by default)
go test ./...

## Reproduce a specific failing order
go test -testcase.seed=1234567890 ./...

## Run multiple times to stress-test independence
go test -count=128 ./...
```

#### Test Independence Rules

Following the schema ensures test independence:

- [ ] Each `Describe` block has its own dedicated `ACT` function
- [ ] Variables use `let.Var` with fresh initialization per test
- [ ] No shared mutable state between tests (use `subject.Super(t)` to inherit, not globals)
- [ ] Cleanup registered via `t.Cleanup()` in `Before` hooks or individual tests
- [ ] Tests pass when run individually AND with `-count=128`

#### Common Ordering Pitfalls

**❌ Anti-Pattern: Shared Global State**

```go
// Don't do this - global variable shared across tests
var counter int

func TestSomething(t *testing.T) {
    s := testcase.NewSpec(t)
    
    s.Then("increments counter", func(t *testcase.T) {
        counter++  // Side effect!
        assert.Equal(t, 1, counter)
    })
    
    s.Then("counter is still valid", func(t *testcase.T) {
        // Depends on previous test's side effect!
        assert.Equal(t, 1, counter)
    })
}
```

**✅ Correct: Isolated State Per Test (Schema Compliant)**

```go
func TestSomething(t *testing.T) {
    s := testcase.NewSpec(t)
    
    counter := let.Var(s, func(t *testcase.T) int {
        return 0  // Fresh for each test
    })
    
    act := func(t *testcase.T) int {
        c := counter.Get(t)
        counter.Set(t, c+1)
        return counter.Get(t)
    }
    
    s.Then("increments counter", func(t *testcase.T) {
        assert.Equal(t, 1, act(t))
    })
    
    s.Then("counter starts fresh", func(t *testcase.T) {
        // Independent - no dependency on other tests
        assert.Equal(t, 0, counter.Get(t))
    })
}
```

**❌ Anti-Pattern: Naming Tests to Control Order**

```go
// Don't do this - relying on lexicographic ordering
func Test01Initialize(t *testing.T) { /* ... */ }
func Test02Process(t *testing.T) { /* depends on 01 */ }
func Test03Cleanup(t *testing.T) { /* depends on 02 */ }
```

**✅ Correct: Independent Tests with Descriptive Names**

```go
// Each test is self-contained, no ordering dependency
func TestInitializeCreatesValidState(t *testing.T) { /* ... */ }
func TestProcessWithValidInputSucceeds(t *testing.T) { /* ... */ }
func TestCleanupReleasesResources(t *testing.T) { /* ... */ }
```

#### Running Tests with Different Strategies

| Command | Purpose |
|---------|---------|
| `go test ./...` | Default random order (testcase default) |
| `go test -count=128 ./...` | Run 128 times to stress-test independence |
| `go test -shuffle=on ./...` | Go's built-in shuffle (optional, testcase does this by default) |
| `go test -testcase.seed=123456 ./...` | Reproduce specific failing order |

#### References

- [Quality Coding: Random Test Order](https://qualitycoding.org/random-test-order/)
- Xcode 10 introduced random test ordering to XCTest for the same reasons
- Simon Whitaker's observation: *"I don't understand why this test keeps failing when run in isolation but always passes as part of the suite..."*

---

### Summary Checklist

Before submitting a test, verify:

- [ ] Test function name matches the system under test
- [ ] Package uses `_test` suffix for black-box testing
- [ ] Happy path tests are at the top level (no `When()` wrapper)
- [ ] Each `Describe` block has a dedicated `ACT` function
- [ ] Variables use `let.Var` with pseudo-random data where appropriate
- [ ] Rainy paths use `When`/`And` to branch from base context
- [ ] Nesting is kept shallow (max 2-3 levels)
- [ ] `Before` hooks are used for shared setup/cleanup
- [ ] Assertions use `assert.Must` for setup, `assert.NoError`/`assert.Equal` for tests

---

*This schema document complements the [TESTCASE.md](./TESTCASE.md) guide. Together they provide a complete reference for writing behavioral tests with the testcase framework.*

## Behavioral Testing with `testcase.NewSpec`

This document provides a comprehensive guide on how to write behavioral tests using the `testcase` framework's `NewSpec` API. The patterns documented here are extracted from real examples across the Frameless codebase.

### Flat Nesting Convention

#### The Golden Rule: Happy Path First, Keep It Flat

By default, aim to keep context nesting flat. Start with the **most simplistic happy path** as the default context. Rainy paths (error cases) are then derived from this base through additional contexts.

**Key Principles:**

1. **Happy path is NOT a separate context** - it should be the default/base test
2. **Rainy paths branch from the base** using `s.When()`, `s.And()`, or separate `s.Test()` calls
3. **Keep nesting shallow** - avoid deeply nested `When`/`Then` chains

#### Example: Flat Structure with Happy Path First

```go
func TestMerge(t *testing.T) {
    s := testcase.NewSpec(t)

    var (
        errs = testcase.Let[[]error](s, nil)
    )
    act := func(t *testcase.T) error {
        return Merge(errs.Get(t)...)
    }

    // Happy path - the default context, no wrapping When()
    s.Test("returns nil when no errors are supplied", func(t *testcase.T) {
        errs.LetValue(s, []error{})
        assert.NoError(t, act(t))
    })

    // Rainy paths branch from base with When/And
    s.When("an error value is supplied", func(s *testcase.Spec) {
        expectedErr := let.Error(s)
        errs.Let(s, func(t *testcase.T) []error {
            return []error{expectedErr.Get(t)}
        })

        s.Then("the exact value is returned", func(t *testcase.T) {
            assert.Equal(t, expectedErr.Get(t), act(t))
        })
    })

    s.And("the error is a typed error", func(s *testcase.Spec) {
        expectedErr.LetValue(s, ErrType1{})

        s.Then("errors.Is can find the wrapped error", func(t *testcase.T) {
            assert.True(t, errors.Is(act(t), ErrType1{}))
        })
    })
}
```

#### Example: File System Happy Path First

```go
func TestFileSystem(t *testing.T) {
    s := testcase.NewSpec(t)

    // Happy path - default context
    s.Test("implements fs.FS", func(t *testcase.T) {
        fsys := localfs.FileSystem{RootPath: t.TempDir()}
        dir := filepath.Join("foo")
        name := filepath.Join(dir, t.Random.UUID())

        assert.NoError(t, fsys.Mkdir(dir, filemode.UserRWX))
        exp := []byte(t.Random.String())

        // Create and write file (happy path)
        infile, err := fsys.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, filemode.UserRWX)
        assert.NoError(t, err)
        t.Cleanup(func() { _ = fsys.Remove(name) })

        _, err = iokit.WriteAll(infile, exp)
        assert.NoError(t, err)
        assert.NoError(t, infile.Close())

        // Read and verify (happy path assertions)
        file1, err := fsys.Open(name)
        assert.NoError(t, err)

        got, err := io.ReadAll(file1)
        assert.NoError(t, file1.Close())
        assert.Equal(t, exp, got)
    })

    // Rainy path - error case derived from base
    s.Test("opening unknown file returns os.ErrNotExist", func(t *testcase.T) {
        fsys := localfs.FileSystem{RootPath: t.TempDir()}

        file, err := fsys.Open("unknown-file-name")
        assert.ErrorIs(t, err, os.ErrNotExist)
        assert.Nil(t, file)
    })
}
```

---

### Basic Structure

#### Creating a Spec

Every behavioral test starts by creating a spec from the standard `*testing.T`:

```go
func TestSomething(t *testing.T) {
    s := testcase.NewSpec(t)

    // Define your tests here
}
```

#### Simple Test Cases (Happy Path Default)

For straightforward tests, use `s.Test()` with a descriptive name. The happy path should be the default - no need to wrap it in `s.When("happy path")`:

```go
func TestMerge(t *testing.T) {
    s := testcase.NewSpec(t)

    // Happy path - default context
    s.Test("returns nil when no errors are supplied", func(t *testcase.T) {
        errs := []error{}
        err := Merge(errs...)
        assert.NoError(t, err)
    })

    // Rainy path - error case
    s.When("errors are supplied", func(s *testcase.Spec) {
        expectedErr := errors.New("boom")

        s.Test("returns the merged error", func(t *testcase.T) {
            errs := []error{expectedErr}
            err := Merge(errs...)
            assert.ErrorIs(t, err, expectedErr)
        })
    })
}
```

---

### Test Organization Patterns

#### BDD-Style with Flat Nesting

Use BDD-style keywords while maintaining flat nesting. The happy path is the default; rainy paths use `s.When()`/`s.And()`:

```go
func TestMerge(t *testing.T) {
    s := testcase.NewSpec(t)

    var (
        errs = testcase.Let[[]error](s, nil)
    )
    act := func(t *testcase.T) error {
        return Merge(errs.Get(t)...)
    }

    // Happy path - default context, no When() wrapper needed
    s.Test("returns nil when no errors are supplied", func(t *testcase.T) {
        errs.LetValue(s, []error{})
        assert.NoError(t, act(t))
    })

    // Rainy paths branch from base with When/And
    s.When("an error value is supplied", func(s *testcase.Spec) {
        expectedErr := let.Error(s)

        errs.Let(s, func(t *testcase.T) []error {
            return []error{expectedErr.Get(t)}
        })

        s.Then("the exact value is returned", func(t *testcase.T) {
            assert.Must(t).Equal(expectedErr.Get(t), act(t))
        })

        s.And("the error is typed", func(s *testcase.Spec) {
            expectedErr.LetValue(s, ErrType1{})

            s.Then("errors.Is finds the wrapped error", func(t *testcase.T) {
                assert.True(t, errors.Is(act(t), ErrType1{}))
            })
        })
    })
}
```

#### Common BDD Keywords and When to Use Them

| Keyword        | Purpose                    | When to Use                                        |
| -------------- | -------------------------- | -------------------------------------------------- |
| `s.Test()`     | Define a test case         | **Happy path** (default) or independent scenarios  |
| `s.Describe()` | Group related tests        | Organizing tests by method/feature                 |
| `s.When()`     | Describe a condition/state | **Rainy paths** - alternative conditions from base |
| `s.And()`      | Add additional conditions  | Further narrowing within a `When` block            |
| `s.Then()`     | Define expected outcome    | Inside `When`/`And` blocks for assertions          |

**Convention:** Use `s.Test()` for the happy path (default context). Use `s.When()` only when you need to change the test setup significantly from the base case.

#### Flat Context Example with Transaction Scenarios

```go
func (spec SpecMemory) SpecAdd(s *testcase.Spec) {
    var (
        event = testcase.Let(s, func(t *testcase.T) interface{} {
            return AddTestEvent{V: `hello world`}
        })
        subject = func(t *testcase.T) error {
            return spec.memoryGet(t).Append(spec.ctxGet(t), event.Get(t))
        }
    )

    // Happy path - default context (background context, no transaction)
    s.Test("appends event successfully", func(t *testcase.T) {
        assert.NoError(t, subject(t))
        assert.Contains(t, spec.memoryGet(t).Events(), event.Get(t))
    })

    // Rainy path - context canceled
    s.When(`context is canceled`, func(s *testcase.Spec) {
        spec.ctx().Let(s, func(t *testcase.T) context.Context {
            c, cancel := context.WithCancel(context.Background())
            cancel()
            return c
        })

        s.Then(`returns with context canceled error`, func(t *testcase.T) {
            assert.Must(t).ErrorIs(context.Canceled, subject(t))
        })
    })

    // Rainy path - during transaction (events buffered until commit)
    s.When(`during transaction`, func(s *testcase.Spec) {
        s.Before(func(t *testcase.T) {
            tx, err := spec.memoryGet(t).BeginTx(spec.ctxGet(t))
            assert.Must(t).NoError(err)
            spec.ctx().Set(t, tx)
        })

        s.Then(`events are buffered until commit`, func(t *testcase.T) {
            assert.NoError(t, subject(t))
            assert.NotContains(t, spec.memoryGet(t).Events(), event.Get(t))

            assert.NoError(t, spec.memoryGet(t).CommitTx(spec.ctxGet(t)))
            assert.Contains(t, spec.memoryGet(t).Events(), event.Get(t))
        })
    })
}
```

---

### Setup and Teardown

#### Using `testcase.Let` for Shared State

`testcase.Let` creates a variable that can be initialized once at the spec level or overridden in nested specs:

```go
func TestMux(t *testing.T) {
    s := testcase.NewSpec(t)

    mux := testcase.Let(s, func(t *testcase.T) *cli.Mux {
        return &cli.Mux{}
    })

    // Access the value in tests
    s.Test("something", func(t *testcase.T) {
        m := mux.Get(t)
        // use m...
    })
}
```

#### Using `testcase.Var` for Custom Initialization

For more control, use `testcase.Var`:

```go
func (spec SpecMemory) memory() testcase.Var[*memory.EventLog] {
    return testcase.Var[*memory.EventLog]{
        ID: `*memory.EventLog`,
        Init: func(t *testcase.T) *memory.EventLog {
            return memory.NewEventLog()
        },
    }
}

func (spec SpecMemory) memoryGet(t *testcase.T) *memory.EventLog {
    return spec.memory().Get(t)
}
```

#### Overriding Values in Nested Specs

```go
s.When("during transaction", func(s *testcase.Spec) {
    s.Before(func(t *testcase.T) {
        tx, err := spec.memoryGet(t).BeginTx(spec.ctxGet(t))
        assert.Must(t).NoError(err)
        spec.ctx().Set(t, tx)  // Override the context variable
    })

    s.Then("behavior in transaction", func(t *testcase.T) {
        // Uses the overridden context
    })
})
```

#### Using `s.Before()` for Hooks

```go
s.Before(func(t *testcase.T) {
    args := slicekit.Clone(request.Get(t).Args)
    t.OnFail(func() {
        t.Log("args:", args)
        t.Log("code:", response.Get(t).Code)
        t.Log("\nout:\n", response.Get(t).Out.String())
        t.Log("\nerr:\n", response.Get(t).Err.String())
    })
})
```

#### Using `t.Cleanup()` for Teardown

```go
s.Test("implements fs.FS", func(t *testcase.T) {
    fsys := localfs.FileSystem{RootPath: t.TempDir()}

    name := filepath.Join(dir, t.Random.UUID())

    infile, err := fsys.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, filemode.UserRWX)
    assert.NoError(t, err)
    t.Cleanup(func() { _ = fsys.Remove(name) })  // Cleanup on test end

    _, err = iokit.WriteAll(infile, exp)
    assert.NoError(t, err)
})
```

---

### Subject Function Pattern

Encapsulate the code under test in a `subject` function for cleaner tests:

```go
func TestMerge(t *testing.T) {
    s := testcase.NewSpec(t)

    var (
        errs = testcase.Let[[]error](s, nil)
    )

    // Subject function - the code under test
    act := func(t *testcase.T) error {
        return errorkitlite.Merge(errs.Get(t)...)
    }

    s.When("no error is supplied", func(s *testcase.Spec) {
        errs.Let(s, func(t *testcase.T) []error {
            return []error{}
        })

        s.Then("it will return with nil", func(t *testcase.T) {
            assert.Must(t).NoError(act(t))  // Call subject through act()
        })
    })
}
```

#### Subject with Multiple Parameters

```go
func (spec SpecMemory) SpecAdd(s *testcase.Spec) {
    type AddTestEvent struct{ V string }
    var (
        event = testcase.Let(s, func(t *testcase.T) interface{} {
            return AddTestEvent{V: `hello world`}
        })

        // Subject encapsulates the method call with all parameters
        subject = func(t *testcase.T) error {
            return spec.memoryGet(t).Append(spec.ctxGet(t), event.Get(t))
        }
    )

    s.When(`context is canceled`, func(s *testcase.Spec) {
        // Override context for this scenario
        spec.ctx().Let(s, func(t *testcase.T) context.Context {
            c, cancel := context.WithCancel(context.Background())
            cancel()
            return c
        })

        s.Then(`returns with context canceled error`, func(t *testcase.T) {
            assert.Must(t).ErrorIs(context.Canceled, subject(t))
        })
    })
}
```

---

### Context Management

#### Defining Context Variables

```go
func (spec SpecMemory) ctx() testcase.Var[context.Context] {
    return testcase.Var[context.Context]{
        ID: `context.Context`,
        Init: func(t *testcase.T) context.Context {
            return context.Background()
        },
    }
}

func (spec SpecMemory) ctxGet(t *testcase.T) context.Context {
    return spec.ctx().Get(t).(context.Context)
}
```

#### Overriding Context in Test Scenarios

```go
s.When(`context is canceled`, func(s *testcase.Spec) {
    spec.ctx().Let(s, func(t *testcase.T) context.Context {
        c, cancel := context.WithCancel(context.Background())
        cancel()  // Cancel immediately
        return c
    })

    s.Then(`returns with context canceled error`, func(t *testcase.T) {
        assert.Must(t).ErrorIs(context.Canceled, subject(t))
    })
})
```

#### Context in Transaction Tests

```go
s.When(`during transaction`, func(s *testcase.Spec) {
    s.Before(func(t *testcase.T) {
        tx, err := spec.memoryGet(t).BeginTx(spec.ctxGet(t))
        assert.Must(t).NoError(err)
        spec.ctx().Set(t, tx)  // Replace context with transaction
    })

    s.Then(`behavior in transaction scope`, func(t *testcase.T) {
        assert.NoError(t, subject(t))
        // Verify transactional behavior
        assert.NotContains(t, spec.memoryGet(t).Events(), event.Get(t))

        // Commit and verify
        assert.NoError(t, spec.memoryGet(t).CommitTx(spec.ctxGet(t)))
        assert.Contains(t, spec.memoryGet(t).Events(), event.Get(t))
    })
})
```

---

### Contract Testing

#### Reusable Spec Structures

Create reusable spec structures for contract testing:

```go
type ConnectionContract struct {
    MakeSubject  func(tb testing.TB) postgresql.Connection
    MakeContext  func(testing.TB) context.Context
    CreateTable  func(ctx context.Context, connection postgresql.Connection, name string) error
    DeleteTable  func(ctx context.Context, connection postgresql.Connection, name string) error
    HasTable     func(ctx context.Context, connection postgresql.Connection, name string) (bool, error)
}

func (c ConnectionContract) cm() testcase.Var[postgresql.Connection] {
    return testcase.Var[postgresql.Connection]{
        ID: "Connection",
        Init: func(t *testcase.T) postgresql.Connection {
            return c.MakeSubject(t)
        },
    }
}

func (c ConnectionContract) Spec(s *testcase.Spec) {
    s.Test(`.BeginTx = transaction`, func(t *testcase.T) {
        p := c.cm().Get(t)

        tx, err := p.BeginTx(c.MakeContext(t))
        assert.NoError(t, err)
        t.Defer(p.RollbackTx, tx)

        name := c.makeTestTableName()
        assert.NoError(t, c.CreateTable(tx, p, name))
        defer c.cleanupTable(t, name)

        assert.NoError(t, p.RollbackTx(tx))

        ctx := c.MakeContext(t)
        has, err := c.HasTable(ctx, p, name)
        assert.NoError(t, err)
        assert.False(t, has)
    })
}
```

#### Running Contract Tests

```go
func TestConnection_PoolContract(t *testing.T) {
    testcase.RunSuite(t, ConnectionContract{
        MakeSubject: func(tb testing.TB) postgresql.Connection {
            cm, err := postgresql.Connect(DatabaseURL(tb))
            assert.NoError(tb, err)
            return cm
        },
        DriverName: "postgres",
        MakeContext: func(tb testing.TB) context.Context {
            return context.Background()
        },
        CreateTable: func(ctx context.Context, connection postgresql.Connection, name string) error {
            _, err := connection.ExecContext(ctx, fmt.Sprintf(`CREATE TABLE %q ();`, name))
            return err
        },
        DeleteTable: func(ctx context.Context, connection postgresql.Connection, name string) error {
            _, err := connection.ExecContext(ctx, fmt.Sprintf(`DROP TABLE %q;`, name))
            return err
        },
        HasTable: func(ctx context.Context, connection postgresql.Connection, name string) (bool, error) {
            var has bool
            query := fmt.Sprintf(`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = '%s') AS e;`, name)
            err := connection.QueryRowContext(ctx, query).Scan(&has)
            return has, err
        },
    })
}
```

#### Using Contract Helpers

```go
func TestRepository(t *testing.T) {
    m := memory.NewMemory()
    repo := memory.NewRepository[TestEntity, string](m)

    testcase.RunSuite(t, resource.Contract[TestEntity, string](repo, resource.Config[TestEntity, string]{
        MetaAccessor:  m,
        CommitManager: m,
        CRUD: crudcontract.Config[TestEntity, string]{
            MakeEntity: makeTestEntity,
        },
    }))
}
```

---

### Random Execution Order

#### Why Random Order Matters

The `testcase` framework executes tests in a **random order by default**. This is intentional and critical for test quality:

- **Flaws hidden by ordering**: Tests that pass in one order may fail in another if they share mutable state or have implicit dependencies
- **Independence enforcement**: Random execution forces you to write truly independent tests with no assumptions about execution sequence
- **Early detection of coupling**: Tests that depend on each other will fail quickly and visibly when shuffled

#### The TESTCASE_SEED

Each spec has a `TESTCASE_SEED` that ensures:

1. **Reproducible failures**: When a test fails due to ordering issues, the seed is printed in the output
2. **Debug with same order**: Re-run with `-testcase.seed=<value>` to reproduce the exact failing sequence
3. **Confidence in fixes**: Verify your fix works by running with the same seed that caused the failure

```bash
## Run tests (random order by default)
go test ./...

## Reproduce a specific failing order
go test -testcase.seed=1234567890 ./...

## Run multiple times to stress-test independence
go test -count=128 ./...
```

#### Test Independence Checklist

Before considering your tests complete, verify:

- [ ] No test relies on side effects from another test
- [ ] Each test cleans up after itself (use `t.Cleanup()`)
- [ ] No shared mutable state between tests
- [ ] Tests pass when run individually AND in random order
- [ ] Running with `-count=128` doesn't reveal ordering-dependent failures

#### Common Ordering Pitfalls

**❌ Anti-Pattern: Shared Global State**

```go
// Don't do this - global variable shared across tests
var counter int

func TestSomething(t *testing.T) {
    s := testcase.NewSpec(t)

    s.Test("increments counter", func(t *testcase.T) {
        counter++  // Side effect!
        assert.Equal(t, 1, counter)
    })

    s.Test("counter is still valid", func(t *testcase.T) {
        // Depends on previous test's side effect!
        assert.Equal(t, 1, counter)
    })
}
```

**✅ Correct: Isolated State Per Test**

```go
func TestSomething(t *testing.T) {
    s := testcase.NewSpec(t)

    counter := let.Var(s, func(t *testcase.T) int {
        return 0  // Fresh for each test
    })

    s.Test("increments counter", func(t *testcase.T) {
        counter.Set(t, counter.Get(t)+1)
        assert.Equal(t, 1, counter.Get(t))
    })

    s.Test("counter starts fresh", func(t *testcase.T) {
        // Independent - no dependency on other tests
        assert.Equal(t, 0, counter.Get(t))
    })
}
```

**❌ Anti-Pattern: Naming Tests to Control Order**

```go
// Don't do this - relying on lexicographic ordering
func Test01Initialize(t *testing.T) { /* ... */ }
func Test02Process(t *testing.T) { /* depends on 01 */ }
func Test03Cleanup(t *testing.T) { /* depends on 02 */ }
```

**✅ Correct: Independent Tests with Descriptive Names**

```go
// Each test is self-contained
func TestInitialize(t *testing.T) { /* ... */ }
func TestProcessWithValidInput(t *testing.T) { /* ... */ }
func TestCleanupReleasesResources(t *testing.T) { /* ... */ }
```

#### Running Tests with Different Strategies

| Command                               | Purpose                                                         |
| ------------------------------------- | --------------------------------------------------------------- |
| `go test ./...`                       | Default random order (testcase default)                         |
| `go test -count=128 ./...`            | Run 128 times to stress-test independence                       |
| `go test -shuffle=on ./...`           | Go's built-in shuffle (optional, testcase does this by default) |
| `go test -testcase.seed=123456 ./...` | Reproduce specific failing order                                |

#### References

- [Quality Coding: Random Test Order](https://qualitycoding.org/random-test-order/)
- Xcode 10 introduced random test ordering to XCTest for the same reasons
- Simon Whitaker's observation: _"I don't understand why this test keeps failing when run in isolation but always passes as part of the suite..."_

---

### Best Practices

#### 1. Happy Path First, Rainy Paths Second

Always start with the happy path as the default context. Use `s.When()` only for rainy paths that require different setup:

```go
// Good - happy path first, then rainy paths
s.Test("appends event successfully", func(t *testcase.T) {
    assert.NoError(t, subject(t))
    assert.Contains(t, events, expectedEvent)
})

s.When("context is canceled", func(s *testcase.Spec) {
    s.Then("returns context canceled error", func(t *testcase.T) {
        assert.ErrorIs(t, subject(t), context.Canceled)
    })
})

// Less clear - happy path wrapped in When()
s.When("happy path", func(s *testcase.Spec) {
    s.Test("appends event successfully", func(t *testcase.T) {
        // ...
    })
})
```

#### 2. Use `s.Test()` for Happy Path, `s.When()` for Variations

The happy path should use the default variable values. Override only when testing variations:

```go
// Default setup - happy path uses background context
spec.ctx().Bind(s)  // defaults to context.Background()

s.Test("appends successfully", func(t *testcase.T) {
    // Uses default background context
    assert.NoError(t, subject(t))
})

// Override only in rainy path scenarios
s.When("during transaction", func(s *testcase.Spec) {
    s.Before(func(t *testcase.T) {
        tx, err := spec.memoryGet(t).BeginTx(spec.ctxGet(t))
        assert.Must(t).NoError(err)
        spec.ctx().Set(t, tx)  // Override only here
    })

    s.Then("events buffered until commit", func(t *testcase.T) {
        // Uses transaction context
    })
})
```

#### 3. Keep Nesting Flat - Avoid Deep `When`/`And` Chains

Deep nesting makes tests hard to read. Prefer separate `s.Test()` calls or shallow `When`/`And`:

```go
// Avoid - deeply nested
s.When("condition A", func(s *testcase.Spec) {
    s.And("and condition B", func(s *testcase.Spec) {
        s.And("but also condition C", func(s *testcase.Spec) {
            s.Then("expectation", func(t *testcase.T) {
                // ...
            })
        })
    })
})

// Prefer - flat structure with separate tests
s.Test("with condition A and B and C", func(t *testcase.T) {
    // Setup for A, B, C
    // Assertions
})

s.When("condition A fails", func(s *testcase.Spec) {
    s.Then("returns error", func(t *testcase.T) {
        // ...
    })
})
```

#### 4. Use Helper Functions for Repetitive Setup

Extract repetitive setup into helper functions to keep tests readable:

```go
s.Test("implements fs.FS", func(t *testcase.T) {
    fsys := localfs.FileSystem{RootPath: t.TempDir()}

    name := filepath.Join(dir, t.Random.UUID())  // Unique UUID
    exp := []byte(t.Random.String())             // Random content

    // ... test with random data
})
```

#### 5. Leverage `t.Random` for Test Data

When setup failures should stop the test:

```go
s.Before(func(t *testcase.T) {
    tx, err := spec.memoryGet(t).BeginTx(spec.ctxGet(t))
    assert.Must(t).NoError(err)  // Stops test on error
    spec.ctx().Set(t, tx)
})
```

The `testcase.T` provides a random data generator:

For assertions within the test:

```go
s.Then("behavior", func(t *testcase.T) {
    assert.NoError(t, subject(t))
    assert.Contains(t, events, expectedEvent)
})
```

#### 6. Use `assert.Must` for Setup Assertions

Provide context when tests fail:

```go
s.Before(func(t *testcase.T) {
    args := slicekit.Clone(request.Get(t).Args)
    t.OnFail(func() {
        t.Log("args:", args)
        t.Log("code:", response.Get(t).Code)
        t.Log("\nout:\n", response.Get(t).Out.String())
        t.Log("\nerr:\n", response.Get(t).Err.String())
    })
})
```

When setup failures should stop the test immediately:

For complex specs, extract the spec logic into methods:

```go
type SpecMemory struct{}

func (spec SpecMemory) Spec(s *testcase.Spec) {
    spec.ctx().Bind(s)
    spec.memory().Bind(s)
    s.Describe(`.Add`, spec.SpecAdd)
}

func (spec SpecMemory) memory() testcase.Var[*memory.EventLog] {
    return testcase.Var[*memory.EventLog]{
        ID: `*memory.EventLog`,
        Init: func(t *testcase.T) *memory.EventLog {
            return memory.NewEventLog()
        },
    }
}

func (spec SpecMemory) SpecAdd(s *testcase.Spec) {
    // Test logic for .Add method
}
```

#### 7. Use `assert.NoError` for Test Assertions

```go
s.Test(`.BeginTx = transaction`, func(t *testcase.T) {
    p := c.cm().Get(t)

    tx, err := p.BeginTx(c.MakeContext(t))
    assert.NoError(t, err)
    t.Defer(p.RollbackTx, tx)  // Ensures rollback even if test fails

    name := c.makeTestTableName()
    assert.NoError(t, c.CreateTable(tx, p, name))
    defer c.cleanupTable(t, name)

    // ... assertions
})
```

For assertions within the test:

Cover both happy and unhappy paths:

```go
s.When("no error is supplied", func(s *testcase.Spec) {
    errs.Let(s, func(t *testcase.T) []error { return []error{} })

    s.Then("it will return with nil", func(t *testcase.T) {
        assert.Must(t).NoError(act(t))
    })
})

s.When("an error value is supplied", func(s *testcase.Spec) {
    expectedErr := let.Error(s)
    errs.Let(s, func(t *testcase.T) []error { return []error{expectedErr.Get(t)} })

    s.Then("the exact value is returned", func(t *testcase.T) {
        assert.Must(t).Equal(expectedErr.Get(t), act(t))
    })
})

s.And("but the error value is nil", func(s *testcase.Spec) {
    expectedErr.LetValue(s, nil)

    s.Then("it will return with nil", func(t *testcase.T) {
        assert.Must(t).NoError(act(t))
    })
})
```

---

### Quick Reference Template

Here's a template you can use to start writing behavioral tests:

```go
func TestYourFeature(t *testing.T) {
    s := testcase.NewSpec(t)

    // Define shared variables at spec level
    var (
        input = testcase.Let(s, func(t *testcase.T) YourType {
            return makeValidInput(t)  // Default to happy path data
        })

        subject = func(t *testcase.T) YourResult {
            return YourFunction(input.Get(t))
        }
    )

    // Happy path - default context (no When() wrapper needed)
    s.Test("returns expected result with valid input", func(t *testcase.T) {
        got := subject(t)
        assert.Equal(t, expectedResult, got)
    })

    // Rainy paths - variations from base using When/And
    s.When("input is nil", func(s *testcase.Spec) {
        input.LetValue(s, nil)

        s.Then("returns error", func(t *testcase.T) {
            assert.Error(t, subject(t))
        })
    })

    s.When("context is canceled", func(s *testcase.Spec) {
        // Override context if your function accepts it
        spec.ctx().Let(s, func(t *testcase.T) context.Context {
            c, cancel := context.WithCancel(context.Background())
            cancel()
            return c
        })

        s.Then("returns context canceled error", func(t *testcase.T) {
            assert.ErrorIs(t, subject(t), context.Canceled)
        })
    })
}
```

---

### Additional Resources

- [`testcase` package documentation](https://pkg.go.dev/go.llib.dev/testcase)
- [`assert` package documentation](https://pkg.go.dev/go.llib.dev/testcase/assert)
- [`let` package for test data generation](https://pkg.go.dev/go.llib.dev/testcase/let)

