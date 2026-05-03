# Behavioral Testing with `testcase.NewSpec`

This document provides a comprehensive guide on how to write behavioral tests using the `testcase` framework's `NewSpec` API. The patterns documented here are extracted from real examples across the Frameless codebase.

## Flat Nesting Convention

### The Golden Rule: Happy Path First, Keep It Flat

By default, aim to keep context nesting flat. Start with the **most simplistic happy path** as the default context. Rainy paths (error cases) are then derived from this base through additional contexts.

**Key Principles:**

1. **Happy path is NOT a separate context** - it should be the default/base test
2. **Rainy paths branch from the base** using `s.When()`, `s.And()`, or separate `s.Test()` calls
3. **Keep nesting shallow** - avoid deeply nested `When`/`Then` chains

### Example: Flat Structure with Happy Path First

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

### Example: File System Happy Path First

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

## Basic Structure

### Creating a Spec

Every behavioral test starts by creating a spec from the standard `*testing.T`:

```go
func TestSomething(t *testing.T) {
    s := testcase.NewSpec(t)

    // Define your tests here
}
```

### Simple Test Cases (Happy Path Default)

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

## Test Organization Patterns

### BDD-Style with Flat Nesting

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

### Common BDD Keywords and When to Use Them

| Keyword        | Purpose                    | When to Use                                        |
| -------------- | -------------------------- | -------------------------------------------------- |
| `s.Test()`     | Define a test case         | **Happy path** (default) or independent scenarios  |
| `s.Describe()` | Group related tests        | Organizing tests by method/feature                 |
| `s.When()`     | Describe a condition/state | **Rainy paths** - alternative conditions from base |
| `s.And()`      | Add additional conditions  | Further narrowing within a `When` block            |
| `s.Then()`     | Define expected outcome    | Inside `When`/`And` blocks for assertions          |

**Convention:** Use `s.Test()` for the happy path (default context). Use `s.When()` only when you need to change the test setup significantly from the base case.

### Flat Context Example with Transaction Scenarios

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

## Setup and Teardown

### Using `testcase.Let` for Shared State

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

### Using `testcase.Var` for Custom Initialization

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

### Overriding Values in Nested Specs

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

### Using `s.Before()` for Hooks

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

### Using `t.Cleanup()` for Teardown

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

## Subject Function Pattern

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

### Subject with Multiple Parameters

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

## Context Management

### Defining Context Variables

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

### Overriding Context in Test Scenarios

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

### Context in Transaction Tests

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

## Contract Testing

### Reusable Spec Structures

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

### Running Contract Tests

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

### Using Contract Helpers

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

### Test Independence Checklist

Before considering your tests complete, verify:

- [ ] No test relies on side effects from another test
- [ ] Each test cleans up after itself (use `t.Cleanup()`)
- [ ] No shared mutable state between tests
- [ ] Tests pass when run individually AND in random order
- [ ] Running with `-count=128` doesn't reveal ordering-dependent failures

### Common Ordering Pitfalls

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

### Running Tests with Different Strategies

| Command                               | Purpose                                                         |
| ------------------------------------- | --------------------------------------------------------------- |
| `go test ./...`                       | Default random order (testcase default)                         |
| `go test -count=128 ./...`            | Run 128 times to stress-test independence                       |
| `go test -shuffle=on ./...`           | Go's built-in shuffle (optional, testcase does this by default) |
| `go test -testcase.seed=123456 ./...` | Reproduce specific failing order                                |

### References

- [Quality Coding: Random Test Order](https://qualitycoding.org/random-test-order/)
- Xcode 10 introduced random test ordering to XCTest for the same reasons
- Simon Whitaker's observation: _"I don't understand why this test keeps failing when run in isolation but always passes as part of the suite..."_

---

## Best Practices

### 1. Happy Path First, Rainy Paths Second

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

### 2. Use `s.Test()` for Happy Path, `s.When()` for Variations

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

### 3. Keep Nesting Flat - Avoid Deep `When`/`And` Chains

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

### 4. Use Helper Functions for Repetitive Setup

Extract repetitive setup into helper functions to keep tests readable:

```go
s.Test("implements fs.FS", func(t *testcase.T) {
    fsys := localfs.FileSystem{RootPath: t.TempDir()}

    name := filepath.Join(dir, t.Random.UUID())  // Unique UUID
    exp := []byte(t.Random.String())             // Random content

    // ... test with random data
})
```

### 5. Leverage `t.Random` for Test Data

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

### 6. Use `assert.Must` for Setup Assertions

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

### 7. Use `assert.NoError` for Test Assertions

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

## Quick Reference Template

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

## Additional Resources

- [`testcase` package documentation](https://pkg.go.dev/go.llib.dev/testcase)
- [`assert` package documentation](https://pkg.go.dev/go.llib.dev/testcase/assert)
- [`let` package for test data generation](https://pkg.go.dev/go.llib.dev/testcase/let)
