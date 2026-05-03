# Spec Definition Hard Rules

This document defines the hard rules that must be followed when defining correct `testcase` specs. These rules are derived from the `testcase` framework's philosophy and best practices.

## Test Function Naming

### Rule 1: The test function name MUST be the testing subject's name

```go
// ✅ Correct
func TestMyType(t *testing.T) { ... }
func TestParseInput(t *testing.T) { ... }
func TestRepository(t *testing.T) { ... }

// ❌ Incorrect
func TestSomething(t *testing.T) { ... }
func TestMainLogic(t *testing.T) { ... }
```

### Rule 2: Test functions MUST follow the `Test<Subject>` convention

- Use PascalCase for the subject name
- Do not prefix with "Test" more than once
- Keep it concise and descriptive

---

## Spec Declaration

### Rule 3: Create spec immediately in test function body

```go
func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)
    
    // ... spec definition
}
```

### Rule 4: Do not defer spec creation

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

## Subject Definition

### Rule 5: Define the System Under Test (SUT) at spec root level

```go
func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)
    
    subject := let.Var(s, func(t *testcase.T) MyType {
        return MyType{}
    })
}
```

### Rule 6: Subject MUST be accessible via getter function or `.Get(t)`

```go
// Using Var
subject := let.Var(s, func(t *testcase.T) MyType { ... })
// Access: subject.Get(t)

// Or helper function
func subject(t *testcase.T) MyType { ... }
```

---

## Describe Blocks

### Rule 7: A `Spec#Describe` MUST have a dedicated `act` defined as a function

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

### Rule 8: Each `Describe` block MUST have exactly one ACT

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

### Rule 9: Describe blocks MUST group tests by method or feature

```go
s.Describe("#MyFunc", func(s *testcase.Spec) { ... })
s.Describe("#Save", func(s *testcase.Spec) { ... })
s.Describe("validation", func(s *testcase.Spec) { ... })
```

---

## Variables and Inputs

### Rule 10: Test variables MUST be defined using `let.Var` or `let.Let`

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

### Rule 11: Variables MUST be accessed via `.Get(t)` in test functions

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

### Rule 12: Use `t.Random` for property testing to enable reproducible failures

```go
input := t.Random.String()           // Random string
id := t.Random.UUID()                // Random UUID  
num := t.Random.IntBetween(1, 100)   // Random int in range
data := t.Random.Bytes(32)           // Random bytes
```

---

## The ACT Pattern

### Rule 13: `act` MUST be defined at the top of each `Describe` block

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

### Rule 14: `act` MUST be immutable - no side effects in test functions

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

### Rule 15: `act` MUST NOT vary within tests of the same Describe block

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

## Then Clauses (Happy Path)

### Rule 16: `Spec#Then` / `Spec#Test` should use the current scope's `act`

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

### Rule 17: Happy path tests MUST NOT be wrapped in `When` blocks

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

### Rule 18: `Then` descriptions MUST express the expected behavior

```go
s.Then("returns the input value unchanged", func(t *testcase.T) { ... })
s.Then("creates a new entity with generated ID", func(t *testcase.T) { ... })
s.Then("persists the entity to storage", func(t *testcase.T) { ... })

// ❌ Avoid vague descriptions
s.Then("works correctly", func(t *testcase.T) { ... })
s.Then("does something", func(t *testcase.T) { ... })
```

### Rule 19: Happy path tests MUST come first in Describe blocks

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

## When/And Contexts (Rainy Paths)

### Rule 20: A `Spec#Context` or `#When`/`#And` MUST have a test arrangement

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

### Rule 21: `When` blocks MUST start with variable modification or hook definition

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

### Rule 22: `And` blocks MUST add additional conditions within a `When` context

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

### Rule 23: `When`/`And` nesting MUST remain shallow (max 2-3 levels)

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

## Before Hooks

### Rule 24: `s.Before()` hooks MUST be defined before test assertions

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

### Rule 25: Before hooks MUST be idempotent and reversible

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

## Nesting Structure

### Rule 26: Keep nesting flat - happy path is the default context

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

### Rule 27: Use `s.Test()` for independent scenarios, `s.Then()` within contexts

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

## Test Independence

### Rule 28: Tests MUST NOT depend on execution order

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

### Rule 29: Tests MUST NOT use numbered naming conventions to imply order

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

### Rule 30: Each test MUST set up its own required state

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

## Summary Checklist

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
