# testcase Spec Schema

A formal specification describing how behavioral tests should be structured using the `testcase` framework. This document defines conventions, patterns, and rules that ensure consistency across the codebase.

---

## Table of Contents

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

## Philosophy

### Black-Box Testing First

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

### Happy Path First, Keep It Flat

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

## Test Function Naming

### Rule: The name MUST be the testing subject's name

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

## Spec Declaration

Every test function starts by creating a spec from `*testing.T`:

```go
func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)
    
    // Define variables, describe blocks, etc.
}
```

---

## Subject Definition

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

### Subject Variations

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

## Describe Blocks

### Purpose

`Describe` blocks group related tests around a specific method or feature:

```go
s.Describe("#MyFunc", func(s *testcase.Spec) {
    // Tests for MyFunc go here
})
```

### Rules for Describe Blocks

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

## Variables and Inputs

### Using `let.Var` for Test Variables

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

### Using `t.Random` for Property Testing

`t.Random` provides pseudo-deterministic random data:

```go
input := t.Random.String()           // Random string
id := t.Random.UUID()                // Random UUID  
num := t.Random.IntBetween(1, 100)   // Random int in range
data := t.Random.Bytes(32)           // Random bytes
```

**Benefit:** If a test fails due to an unhandled random input, you get the `TESTCASE_SEED` that can recreate the failing scenario 1:1.

### Variable Access Patterns

| Method | Purpose | Example |
|--------|---------|---------|
| `.Get(t)` | Get current value | `input.Get(t)` |
| `.Set(t, v)` | Set value in this spec | `input.Set(t, "value")` |
| `.Let(s, fn)` | Override in nested spec | `input.Let(s, func(t) { ... })` |
| `.LetValue(s, v)` | Override with constant | `input.LetValue(s, nil)` |
| `.Super(t)` | Get parent value | `subject.Super(t)` |

---

## The ACT Pattern

### Definition

The **ACT** is an immutable testing function that represents the action being tested:

```go
act := func(t *testcase.T) string {
    return subject.Get(t).MyFunc(input.Get(t))
}
```

### Rules for ACT

1. **Must be defined at the top of each `Describe` block**
2. **Must be immutable** - no side effects in test functions themselves
3. **Centralizes arrangement** - forces you to arrange inputs and context upfront
4. **Reduces mental model complexity** - tests focus on context building, not ACT construction

### Why the ACT Pattern Matters

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

## Then Clauses (Happy Path)

### Placement

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

### Naming Convention

`Then` descriptions should express the expected behavior:

```go
s.Then("returns the input value unchanged", func(t *testcase.T) { ... })
s.Then("creates a new entity with generated ID", func(t *testcase.T) { ... })
s.Then("persists the entity to storage", func(t *testcase.T) { ... })
```

---

## When/And Contexts (Rainy Paths)

### Purpose

`When` and `And` blocks describe alternative conditions or states that differ from the happy path:

```go
s.When("context is canceled", func(s *testcase.Spec) {
    // Arrange context for this scenario
    
    s.Then("returns context.Canceled error", func(t *testcase.T) {
        // Assert expected behavior
    })
})
```

### Rules for When/And

1. **Must start with arrangement** - modify variables or set up hooks
2. **Can nest `And` blocks** for additional conditions (but keep it shallow!)
3. **Inherits parent context** - all parent variables and hooks are available

### Variable Modification in Contexts

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

### Using `And` for Additional Conditions

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

## Before Hooks

### Purpose

`s.Before()` hooks run before each test in the context. Use them for:
- Setup that applies to multiple tests
- Cleanup registration via `t.Cleanup()`
- Debug logging on failure via `t.OnFail()`

### Hook Execution Order

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

### Common Before Hook Patterns

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

## Complete Example

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

### System Under Test

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

## Common Patterns

### Pattern 1: Testing Context-Aware Functions

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

### Pattern 2: Testing with Transactions

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

### Pattern 3: Testing Error Cases with Nested Conditions

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

## Anti-Patterns to Avoid

### ❌ Anti-Pattern: Wrapping Happy Path in `When`

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

### ❌ Anti-Pattern: Deep Nesting

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

### ❌ Anti-Pattern: Varying ACT Within Tests

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

### ❌ Anti-Pattern: Missing ACT Definition

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

## Random Execution Order

### Why Random Order Matters

The `testcase` framework executes tests in a **random order by default**. This is intentional and critical for test quality:

- **Flaws hidden by ordering**: Tests that pass in one order may fail in another if they share mutable state or have implicit dependencies
- **Independence enforcement**: Random execution forces you to write truly independent tests with no assumptions about execution sequence
- **Early detection of coupling**: Tests that depend on each other will fail quickly and visibly when shuffled

### The TESTCASE_SEED

Each spec has a `TESTCASE_SEED` that ensures:

1. **Reproducible failures**: When a test fails due to ordering issues, the seed is printed in the output
2. **Debug with same order**: Re-run with `-testcase.seed=<value>` to reproduce the exact failing sequence
3. **Confidence in fixes**: Verify your fix works by running with the same seed that caused the failure

```bash
# Run tests (random order by default)
go test ./...

# Reproduce a specific failing order
go test -testcase.seed=1234567890 ./...

# Run multiple times to stress-test independence
go test -count=128 ./...
```

### Test Independence Rules

Following the schema ensures test independence:

- [ ] Each `Describe` block has its own dedicated `ACT` function
- [ ] Variables use `let.Var` with fresh initialization per test
- [ ] No shared mutable state between tests (use `subject.Super(t)` to inherit, not globals)
- [ ] Cleanup registered via `t.Cleanup()` in `Before` hooks or individual tests
- [ ] Tests pass when run individually AND with `-count=128`

### Common Ordering Pitfalls

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

### Running Tests with Different Strategies

| Command | Purpose |
|---------|---------|
| `go test ./...` | Default random order (testcase default) |
| `go test -count=128 ./...` | Run 128 times to stress-test independence |
| `go test -shuffle=on ./...` | Go's built-in shuffle (optional, testcase does this by default) |
| `go test -testcase.seed=123456 ./...` | Reproduce specific failing order |

### References

- [Quality Coding: Random Test Order](https://qualitycoding.org/random-test-order/)
- Xcode 10 introduced random test ordering to XCTest for the same reasons
- Simon Whitaker's observation: *"I don't understand why this test keeps failing when run in isolation but always passes as part of the suite..."*

---

## Summary Checklist

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
