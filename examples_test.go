package testcase_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/faultinject"
	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/internal/example/memory"
	"github.com/adamluzsi/testcase/internal/example/mydomain"
	"github.com/adamluzsi/testcase/internal/example/spechelper"
	"github.com/adamluzsi/testcase/random"
)

func ExampleSpec() {
	var tb testing.TB

	// spec do not use any global magic
	// it is just a simple abstraction around testing.T#Context
	// Basically you can easily can run it as you would any other go testCase
	//   -> `go run ./... -v -run "my/edge/case/nested/block/I/want/to/run/only"`
	//
	spec := testcase.NewSpec(tb)

	// when you have no side effects in your testing suite,
	// you can enable parallel execution.
	// You can play parallel even from nested specs to apply parallel testing for that spec and below.
	spec.Parallel()
	// or
	spec.NoSideEffect()

	// testcase.variables are thread safe way of setting up complex contexts
	// where some variable need to have different values for edge cases.
	// and I usually work with in-memory implementation for certain shared specs,
	// to make my testCase coverage run fast and still close to somewhat reality in terms of integration.
	// and to me, it is a necessary thing to have "T#parallel" SpecOption safely available
	var myType = func(t *testcase.T) *mydomain.MyUseCase {
		return &mydomain.MyUseCase{}
	}

	// Each describe has a testing subject as an "act" function
	spec.Describe(`IsLower`, func(s *testcase.Spec) {
		var ( // inputs for the Act
			input = testcase.Var[string]{ID: `input`}
		)
		act := func(t *testcase.T) bool {
			return myType(t).IsLower(input.Get(t))
		}

		s.When(`input string has lower case characters`, func(s *testcase.Spec) {
			input.Let(s, func(t *testcase.T) string {
				return t.Random.StringNWithCharset(t.Random.Int(), strings.ToLower(random.CharsetAlpha()))
			})

			s.And(`the first character is capitalized`, func(s *testcase.Spec) {
				// you can add more nesting for more concrete specifications,
				// in each nested block, you work on a separate variable stack,
				// so even if you overwrite something here,
				// that has no effect outside of this scope
				s.Before(func(t *testcase.T) {
					upperCaseLetter := t.Random.StringNC(1, strings.ToUpper(random.CharsetAlpha()))
					input.Set(t, upperCaseLetter+input.Get(t))
				})

				s.Then(`it will report false`, func(t *testcase.T) {
					t.Must.True(act(t),
						fmt.Sprintf(`it was expected that %q will be reported to be not lowercase`, input.Get(t)))
				})

			})

			s.Then(`it will return true`, func(t *testcase.T) {
				t.Must.True(act(t),
					fmt.Sprintf(`it was expected that the %q will re reported to be lowercase`, input.Get(t)))
			})
		})

		s.When(`input string has upcase case characters`, func(s *testcase.Spec) {
			input.Let(s, func(t *testcase.T) string {
				return t.Random.StringNWithCharset(t.Random.Int(), strings.ToUpper(random.CharsetAlpha()))
			})

			s.Then(`it will return false`, func(t *testcase.T) {
				t.Must.False(act(t))
			})
		})
	})
}

func Example_assertWaiterWait() {
	w := assert.Waiter{WaitDuration: time.Millisecond}

	w.Wait() // will wait 1 millisecond and attempt to schedule other go routines
}

func Example_assertWaiterWhile() {
	w := assert.Waiter{
		WaitDuration: time.Millisecond,
		Timeout:      time.Second,
	}

	// will attempt to wait until condition returns false.
	// The maximum time it is willing to wait is equal to the wait timeout duration.
	w.While(func() bool {
		return rand.Intn(1) == 0
	})
}

func ExampleVar() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		storage = testcase.Let[mydomain.Storage](s, func(t *testcase.T) mydomain.Storage {
			return memory.NewStorage()
		})
		subject = testcase.Let(s, func(t *testcase.T) *mydomain.MyUseCase {
			return &mydomain.MyUseCase{Storage: storage.Get(t)}
		})
	)

	s.Describe(`#MyFunc`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			// after GO2 this will be replaced with concrete Types instead of interface{}
			subject.Get(t).MyFunc()
		}

		s.Then(`do some testCase`, func(t *testcase.T) {
			subject(t) // act
			// assertions here.
		})

		// ...
		// other cases with resource xy state change
	})
}

func ExampleVar_Get() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Let(s, func(t *testcase.T) interface{} {
		return 42
	})

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t) // -> 42
	})
}

func ExampleVar_Set() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Let(s, func(t *testcase.T) interface{} {
		return 42
	})

	s.Before(func(t *testcase.T) {
		value.Set(t, 24)
	})

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t) // -> 24
	})
}

func ExampleVar_Let() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Var[int]{
		ID: `the variable group`,
		Init: func(t *testcase.T) int {
			return 42
		},
	}

	value.Let(s, nil)

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t) // -> 42
	})
}

func ExampleVar_Let_valueDefinedAtTestingContextScope() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Var[int]{ID: `the variable group`}

	value.Let(s, func(t *testcase.T) int {
		return 42
	})

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t) // -> 42
	})
}

func ExampleVar_LetValue() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Var[int]{ID: `the variable group`}

	value.LetValue(s, 42)

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t) // -> 42
	})
}

func ExampleVar_EagerLoading() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Let(s, func(t *testcase.T) interface{} {
		return 42
	})

	// will be loaded early on, before the test case block reached.
	// This can be useful when you want to have variables,
	// that also must be present in some sort of attached resource,
	// and as part of the constructor, you want to save it.
	// So when the testCase block is reached, the entity is already present in the resource.
	value.EagerLoading(s)

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t) // -> 42
		// value returned from cache instead of triggering first time initialization.
	})
}

func ExampleVar_Let_eagerLoading() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Var[int]{ID: `value`}

	value.Let(s, func(t *testcase.T) int {
		return 42
	}).EagerLoading(s)

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t) // -> 42
		// value returned from cache instead of triggering first time initialization.
	})
}

func ExampleVar_LetValue_eagerLoading() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Var[int]{ID: `value`}
	value.LetValue(s, 42).EagerLoading(s)

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t) // -> 42
		// value returned from cache instead of triggering first time initialization.
	})
}

func ExampleVar_init() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	value := testcase.Var[int]{
		ID: `value`,
		Init: func(t *testcase.T) int {
			return 42
		},
	}

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t) // 42
	})
}

func ExampleVar_onLet() {
	// package spechelper
	var db = testcase.Var[*sql.DB]{
		ID: `db`,
		Init: func(t *testcase.T) *sql.DB {
			db, err := sql.Open(`driver`, `dataSourceName`)
			if err != nil {
				t.Fatal(err.Error())
			}
			return db
		},
		OnLet: func(s *testcase.Spec, _ testcase.Var[*sql.DB]) {
			s.Tag(`database`)
			s.Sequential()
		},
	}

	var tb testing.TB
	s := testcase.NewSpec(tb)
	db.Let(s, nil)
	s.Test(`some testCase`, func(t *testcase.T) {
		_ = db.Get(t)
		t.HasTag(`database`) // true
	})
}

func ExampleVar_Bind() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	v := testcase.Var[int]{ID: "myvar", Init: func(t *testcase.T) int { return 42 }}
	v.Bind(s)
	s.Test(``, func(t *testcase.T) {
		_ = v.Get(t) // -> 42
	})
}

func ExampleVar_before() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	v := testcase.Var[int]{
		ID:   "myvar",
		Init: func(t *testcase.T) int { return 42 },
		Before: func(t *testcase.T, v testcase.Var[int]) {
			t.Logf(`I'm from the Var.Before block, and the value: %#v`, v.Get(t))
		},
	}
	s.Test(``, func(t *testcase.T) {
		_ = v.Get(t)
		// log: I'm from the Var.Before block
		// -> 42
	})
}

func ExampleT_SkipUntil() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Test(`will be skipped`, func(t *testcase.T) {
		// make tests skip until the given day is reached,
		// then make the tests fail.
		// This helps to commit code which still work in progress.
		t.SkipUntil(2020, 01, 01)
	})

	s.Test(`will not be skipped`, func(t *testcase.T) {})
}

func ExampleT_random() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	s.Test(``, func(t *testcase.T) {
		_ = t.Random.Int()
		_ = t.Random.IntBetween(0, 42)
		_ = t.Random.IntN(42)
		_ = t.Random.Float32()
		_ = t.Random.Float64()
		_ = t.Random.String()
		_ = t.Random.StringN(42)
		_ = t.Random.StringNWithCharset(42, "abc")
		_ = t.Random.Bool()
		_ = t.Random.Time()
		_ = t.Random.TimeN(time.Now(), 0, 4, 2)
		_ = t.Random.TimeBetween(time.Now().Add(-1*time.Hour), time.Now().Add(time.Hour))
		_ = t.Random.ElementFromSlice([]int{1, 2, 3}).(int)
		_ = t.Random.KeyFromMap(map[string]struct{}{`foo`: {}, `bar`: {}, `baz`: {}}).(string)
	})
}

func ExampleT_HasTag() {
	var t *testing.T
	var s = testcase.NewSpec(t)

	type DB interface { // header interface in supplier pkg
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	}
	testcase.Let(s, func(t *testcase.T) DB {
		db, err := sql.Open(`driverName`, `dataSourceName`)
		t.Must.Nil(err)

		if t.HasTag(`black box`) {
			// tests with black box  use http testCase server or similar things and high level tx management not maintainable.
			t.Defer(db.Close)
			return db
		}

		tx, err := db.BeginTx(context.Background(), nil)
		t.Must.Nil(err)
		t.Defer(tx.Rollback)
		return tx
	})
}

func ExampleT_Eventually() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	s.Test(``, func(t *testcase.T) {
		// Eventually this will pass eventually
		t.Eventually(func(it assert.It) {
			it.Must.True(t.Random.Bool())
		})
	})
}

func ExampleT_Defer_withArgs() {
	var t *testing.T
	s := testcase.NewSpec(t)

	something := testcase.Let(s, func(t *testcase.T) *ExampleDeferTeardownWithArgs {
		ptr := &ExampleDeferTeardownWithArgs{}
		// T#Defer arguments copied upon pass by value
		// and then passed to the function during the execution of the deferred function call.
		//
		// This is ideal for situations where you need to guarantee that a value cannot be muta
		t.Defer(ptr.SomeTeardownWithArg, `Hello, World!`)
		return ptr
	})

	s.Test(`a simple test case`, func(t *testcase.T) {
		entity := something.Get(t)

		entity.DoSomething()
	})
}

type ExampleDeferTeardownWithArgs struct{}

func (*ExampleDeferTeardownWithArgs) SomeTeardownWithArg(_ string) {}

func (*ExampleDeferTeardownWithArgs) DoSomething() {}

func ExampleT_Defer() {
	var t *testing.T
	s := testcase.NewSpec(t)

	// db for example is something that needs to defer an action after the testCase run
	db := testcase.Let(s, func(t *testcase.T) *sql.DB {
		db, err := sql.Open(`driverName`, `dataSourceName`)

		// asserting error here with the *testcase.T ensure that the testCase will don't have some spooky failure.
		t.Must.Nil(err)

		// db.Close() will be called after the current test case reach the teardown hooks
		t.Defer(db.Close)

		// check if connection is OK
		t.Must.Nil(db.Ping())

		// return the verified db instance for the caller
		// this db instance will be memorized during the runtime of the test case
		return db
	})

	s.Test(`a simple test case`, func(t *testcase.T) {
		db := db.Get(t)
		t.Must.Nil(db.Ping()) // just to do something with it.
	})
}

func ExampleT_must() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	s.Test(``, func(t *testcase.T) {
		// failed test will stop with FailNow
		t.Must.Equal(1, 1, "must be equal")
	})
}

func ExampleT_should() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	s.Test(``, func(t *testcase.T) {
		// failed test will proceed, but mart the test failed
		t.Should.Equal(1, 1, "should be equal")
	})
}

func ExampleStubTB_testingATestHelper() {
	stub := &doubles.TB{}
	stub.Log("hello", "world")
	fmt.Println(stub.Logs.String())

	myTestHelper := func(tb testing.TB) {
		tb.FailNow()
	}

	var tb testing.TB
	assert.Must(tb).Panic(func() {
		myTestHelper(stub)
	})
	assert.Must(tb).True(stub.IsFailed)
}

func ExampleSpec_withBenchmark() {
	var b *testing.B
	s := testcase.NewSpec(b)

	myType := func(t *testcase.T) *mydomain.MyUseCase {
		return &mydomain.MyUseCase{}
	}

	s.When(`something`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Log(`setup`)
		})

		s.Then(`this benchmark block will be executed by *testing.B.N times`, func(t *testcase.T) {
			myType(t).IsLower(`Hello, World!`)
		})
	})
}

func ExampleSpec_When() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		myType  = func(t *testcase.T) *mydomain.MyUseCase { return &mydomain.MyUseCase{} }
		input   = testcase.Var[string]{ID: `input`}
		subject = func(t *testcase.T) bool { return myType(t).IsLower(input.Get(t)) }
	)

	s.When(`input has only upcase letter`, func(s *testcase.Spec) {
		input.LetValue(s, "UPPER")

		s.Then(`it will be false`, func(t *testcase.T) {
			t.Must.True(!subject(t))
		})
	})

	s.When(`input has only lowercase letter`, func(s *testcase.Spec) {
		input.LetValue(s, "lower")

		s.Then(`it will be true`, func(t *testcase.T) {
			t.Must.True(subject(t))
		})
	})
}

func ExampleSpec_Then() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Then(`it is expected.... so this is the testCase description here`, func(t *testcase.T) {
		// ...
	})
}

func ExampleSpec_Test() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Test(`my testCase description`, func(t *testcase.T) {
		// ...
	})
}

func ExampleSpec_Tag() {
	// example usage:
	// 	TESTCASE_TAG_INCLUDE='E2E' go testCase ./...
	// 	TESTCASE_TAG_EXCLUDE='E2E' go testCase ./...
	// 	TESTCASE_TAG_INCLUDE='E2E' TESTCASE_TAG_EXCLUDE='list,of,excluded,tags' go testCase ./...
	//
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Context(`E2E`, func(s *testcase.Spec) {
		// by tagging the spec spec, we can filter tests orderingOutput later in our CI/CD pipeline.
		// A comma separated list can be set with TESTCASE_TAG_INCLUDE env variable to filter down to tests with certain tags.
		// And/Or a comma separated list can be provided with TESTCASE_TAG_EXCLUDE to exclude tests tagged with certain tags.
		s.Tag(`E2E`)

		s.Test(`some E2E testCase`, func(t *testcase.T) {
			// ...
		})
	})
}

func ExampleSpec_SkipBenchmark() {
	var b *testing.B
	s := testcase.NewSpec(b)
	s.SkipBenchmark()

	s.Test(`this will be skipped during benchmark`, func(t *testcase.T) {})

	s.Context(`some spec`, func(s *testcase.Spec) {
		s.Test(`this as well`, func(t *testcase.T) {})
	})
}

func ExampleSpec_SkipBenchmark_scopedWithContext() {
	var b *testing.B
	s := testcase.NewSpec(b)

	s.When(`rainy path`, func(s *testcase.Spec) {
		s.SkipBenchmark()

		s.Test(`will be skipped during benchmark`, func(t *testcase.T) {})
	})

	s.Context(`happy path`, func(s *testcase.Spec) {
		s.Test(`this will run as benchmark`, func(t *testcase.T) {})
	})
}

func ExampleSpec_Sequential() {
	var t *testing.T
	s := testcase.NewSpec(t)
	s.Sequential() // tells the specs to run list test case in sequence

	s.Test(`this will run in sequence`, func(t *testcase.T) {})

	s.Context(`some spec`, func(s *testcase.Spec) {
		s.Test(`this run in sequence`, func(t *testcase.T) {})

		s.Test(`this run in sequence`, func(t *testcase.T) {})
	})
}

func ExampleSpec_Sequential_scopedWithContext() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Parallel() // on top level, spec marked as parallel

	s.Context(`spec marked sequential`, func(s *testcase.Spec) {
		s.Sequential() // but in subcontext the testCase marked as sequential

		s.Test(`this run in sequence`, func(t *testcase.T) {})
	})

	s.Context(`spec that inherit parallel flag`, func(s *testcase.Spec) {

		s.Test(`this will run in parallel`, func(t *testcase.T) {})
	})
}

func ExampleSpec_HasSideEffect() {
	var t *testing.T
	s := testcase.NewSpec(t)
	// this mark the testCase to contain side effects.
	// this forbids any parallel testCase execution to avoid retry tests.
	//
	// Under the hood this is a syntax sugar for Sequential
	s.HasSideEffect()

	s.Test(`this will run in sequence`, func(t *testcase.T) {})

	s.Context(`some spec`, func(s *testcase.Spec) {
		s.Test(`this run in sequence`, func(t *testcase.T) {})

		s.Test(`this run in sequence`, func(t *testcase.T) {})
	})
}

func ExampleSpec_Sequential_globalVar() {
	var t *testing.T
	s := testcase.NewSpec(t)

	// might or might not allow parallel execution
	// It depends on the
	storage := spechelper.Storage.Bind(s)

	// Tells that the subject of this specification should be software side effect free on its own.
	s.NoSideEffect()

	s.Test("will only run parallel if no dependency has side effect", func(t *testcase.T) {
		t.Logf("%#v", storage.Get(t))
	})
}

func ExampleSpec_Parallel() {
	var t *testing.T
	s := testcase.NewSpec(t)
	s.Parallel() // tells the specs to run list test case in parallel

	s.Test(`this will run in parallel`, func(t *testcase.T) {})

	s.Context(`some spec`, func(s *testcase.Spec) {
		s.Test(`this run in parallel`, func(t *testcase.T) {})

		s.Test(`this run in parallel`, func(t *testcase.T) {})
	})
}

func ExampleSpec_Parallel_scopedWithContext() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Context(`spec marked parallel`, func(s *testcase.Spec) {
		s.Parallel()

		s.Test(`this run in parallel`, func(t *testcase.T) {})
	})

	s.Context(`spec without parallel`, func(s *testcase.Spec) {

		s.Test(`this will run in sequence`, func(t *testcase.T) {})
	})
}

func ExampleSpec_NoSideEffect() {
	var t *testing.T
	s := testcase.NewSpec(t)
	// this is an idiom to express that the subject in the tests here are not expected to have any side-effect.
	// this means they are safe to be executed in parallel.
	s.NoSideEffect()

	s.Test(`this will run in parallel`, func(t *testcase.T) {})

	s.Context(`some spec`, func(s *testcase.Spec) {
		s.Test(`this run in parallel`, func(t *testcase.T) {})

		s.Test(`this run in parallel`, func(t *testcase.T) {})
	})
}

func ExampleSpec_LetValue() {
	var t *testing.T
	s := testcase.NewSpec(t)

	variable := testcase.LetValue(s, "value")

	s.Then(`test case`, func(t *testcase.T) {
		t.Log(variable.Get(t)) // -> "value"
	})
}

func ExampleSpec_LetValue_usageWithinNestedScope() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var myType = func(t *testcase.T) *mydomain.MyUseCase { return &mydomain.MyUseCase{} }

	s.Describe(`#IsLower`, func(s *testcase.Spec) {
		var (
			input   = testcase.Var[string]{ID: `input`}
			subject = func(t *testcase.T) bool {
				return myType(t).IsLower(input.Get(t))
			}
		)

		s.When(`input characters are list lowercase`, func(s *testcase.Spec) {
			testcase.LetValue(s, "list lowercase")
			// or
			input.LetValue(s, "list lowercase")

			s.Then(`it will report true`, func(t *testcase.T) {
				t.Must.True(subject(t))
			})
		})

		s.When(`input is a capitalized`, func(s *testcase.Spec) {
			testcase.LetValue(s, "Capitalized")
			// or
			input.LetValue(s, "Capitalized")

			s.Then(`it will report false`, func(t *testcase.T) {
				t.Must.True(!subject(t))
			})
		})
	})
}

func ExampleSpec_Let() {
	var t *testing.T
	s := testcase.NewSpec(t)

	myTestVar := testcase.Let(s, func(t *testcase.T) interface{} {
		return "value that needs complex construction or can be mutated"
	})

	s.Then(`test case`, func(t *testcase.T) {
		t.Log(myTestVar.Get(t)) // -> returns the value set in the current spec spec for MyTestVar
	})
}

func ExampleSpec_Let_eagerLoading() {
	var t *testing.T
	s := testcase.NewSpec(t)

	myTestVar := testcase.Let(s, func(t *testcase.T) interface{} {
		return "value that will be eager loaded before the testCase/then block reached"
	}).EagerLoading(s)
	// EagerLoading will ensure that the value of this Spec Var will be evaluated during the preparation of the testCase.

	s.Then(`test case`, func(t *testcase.T) {
		t.Log(myTestVar.Get(t))
	})
}

type SupplierWithDBDependency struct {
	DB interface {
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	}
}

func (s SupplierWithDBDependency) DoSomething(ctx context.Context) error {
	rows, err := s.DB.QueryContext(ctx, `SELECT 1 = 1`)
	if err != nil {
		return err
	}
	return rows.Close()
}

func ExampleSpec_Let_sqlDB() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		tx = testcase.Let(s, func(t *testcase.T) *sql.Tx {
			// it is advised to use a persistent db connection between multiple specification runs,
			// because otherwise `go testCase -count $times` can receive random connection failures.
			tx, err := getDBConnection(t).Begin()
			if err != nil {
				t.Fatal(err.Error())
			}
			// testcase.T#Defer will execute the received function after the current testCase edge case
			// where the `tx` testCase variable were accessed.
			t.Defer(tx.Rollback)
			return tx
		})
		supplier = testcase.Let(s, func(t *testcase.T) SupplierWithDBDependency {
			return SupplierWithDBDependency{DB: tx.Get(t)}
		})
	)

	s.Describe(`#DoSomething`, func(s *testcase.Spec) {
		var (
			ctx = testcase.Let(s, func(t *testcase.T) context.Context {
				return context.Background()
			})
			subject = func(t *testcase.T) error {
				return supplier.Get(t).DoSomething(ctx.Get(t))
			}
		)

		s.When(`...`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				//...
			})

			s.Then(`...`, func(t *testcase.T) {
				t.Must.Nil(subject(t))
			})
		})
	})
}

func getDBConnection(_ testing.TB) *sql.DB {
	// logic to retrieve cached db connection in the testing environment
	return nil
}

func ExampleSpec_Let_usageWithinNestedScope() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var myType = func(t *testcase.T) *mydomain.MyUseCase { return &mydomain.MyUseCase{} }

	s.Describe(`#IsLower`, func(s *testcase.Spec) {
		var (
			input   = testcase.Var[string]{ID: `input`}
			subject = func(t *testcase.T) bool {
				return myType(t).IsLower(input.Get(t))
			}
		)

		s.When(`input characters are list lowercase`, func(s *testcase.Spec) {
			testcase.Let(s, func(t *testcase.T) interface{} {
				return "list lowercase"
			})
			// or
			input.Let(s, func(t *testcase.T) string {
				return "list lowercase"
			})

			s.Then(`it will report true`, func(t *testcase.T) {
				t.Must.True(subject(t))
			})
		})

		s.When(`input is a capitalized`, func(s *testcase.Spec) {
			testcase.Let(s, func(t *testcase.T) interface{} {
				return "Capitalized"
			})
			// or
			input.Let(s, func(t *testcase.T) string {
				return "Capitalized"
			})

			s.Then(`it will report false`, func(t *testcase.T) {
				t.Must.True(!subject(t))
			})
		})
	})
}

func ExampleSpec_Let_testingDouble() {
	var t *testing.T
	s := testcase.NewSpec(t)

	stubTB := testcase.Let(s, func(t *testcase.T) *doubles.TB {
		stub := &doubles.TB{}
		t.Defer(stub.Finish)
		return stub
	})

	s.When(`some scope where double should behave in a certain way`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			stubTB.Get(t).StubName = "my stubbed name"
		})

		s.Then(`double will be available in every test case and finishNow called afterwards`, func(t *testcase.T) {
			// ...
		})
	})
}

func ExampleSpec_Before() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Before(func(t *testcase.T) {
		// this will run before the test cases.
	})
}

func ExampleSpec_After() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.After(func(t *testcase.T) {
		// this will run after the test cases.
		// this hook applied to this scope and anything that is nested from here.
		// hooks can be stacked with each call.
	})
}

func ExampleSpec_Around() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Around(func(t *testcase.T) func() {
		// this will run before the test cases

		// this hook applied to this scope and anything that is nested from here.
		// hooks can be stacked with each call
		return func() {
			// The content of the returned func will be deferred to run after the test cases.
		}
	})
}

func ExampleSpec_BeforeAll() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.BeforeAll(func(tb testing.TB) {
		// this will run once before every test cases.
	})
}

func ExampleSpec_AfterAll() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.AfterAll(func(tb testing.TB) {
		// this will run once all the test case already ran.
	})
}

func ExampleSpec_AroundAll() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.AroundAll(func(tb testing.TB) func() {
		// this will run once before all the test case.
		return func() {
			// this will run once after all the test case already ran.
		}
	})
}

func ExampleSpec_Describe() {
	var t *testing.T
	s := testcase.NewSpec(t)

	myType := testcase.Let(s, func(t *testcase.T) *mydomain.MyUseCase {
		return &mydomain.MyUseCase{}
	})

	// Describe description points orderingOutput the subject of the tests
	s.Describe(`#IsLower`, func(s *testcase.Spec) {
		var (
			input   = testcase.Var[string]{ID: `input`}
			subject = func(t *testcase.T) bool {
				// subject should represent what will be tested in the describe block
				return myType.Get(t).IsLower(input.Get(t))
			}
		)

		s.Test(``, func(t *testcase.T) { subject(t) })
	})
}

func ExampleSpec_Context() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Context(`description of the testing spec`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			// prepare for the testing spec
		})

		s.Then(`assert expected outcome`, func(t *testcase.T) {

		})
	})
}

func ExampleSpec_And() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.When(`some spec`, func(s *testcase.Spec) {
		// fulfil the spec

		s.And(`additional spec`, func(s *testcase.Spec) {

			s.Then(`assert`, func(t *testcase.T) {

			})
		})

		s.And(`additional spec opposite`, func(s *testcase.Spec) {

			s.Then(`assert`, func(t *testcase.T) {

			})
		})
	})
}

func ExampleRace() {
	v := mydomain.MyUseCase{}

	// running `go test` with the `-race` flag should help you detect unsafe implementations.
	// each block run at the same time in a race situation
	testcase.Race(func() {
		v.ThreadSafeCall()
	}, func() {
		v.ThreadSafeCall()
	})
}

func ExampleGroup() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Context(`description`, func(s *testcase.Spec) {

		s.Test(``, func(t *testcase.T) {})

	}, testcase.Group(`testing-group-group-that-can-be-even-targeted-with-testCase-run-cli-option`))
}

func ExampleSkipBenchmark() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test(`will run`, func(t *testcase.T) {
		// this will run during benchmark execution
	})

	s.Test(`will skip`, func(t *testcase.T) {
		// this will skip the benchmark execution
	}, testcase.SkipBenchmark())
}

func ExampleFlaky_retryUntilTimeout() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test(`testCase with "random" fails`, func(t *testcase.T) {
		// This testCase might fail "randomly" but the retry flag will allow some tolerance
		// This should be used to find time in team's calendar
		// and then allocate time outside of death-march times to learn to avoid retry tests in the future.
	}, testcase.Flaky(time.Minute))
}

func ExampleFlaky_retryNTimes() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test(`testCase with "random" fails`, func(t *testcase.T) {
		// This testCase might fail "randomly" but the retry flag will allow some tolerance
		// This should be used to find time in team's calendar
		// and then allocate time outside of death-march times to learn to avoid retry tests in the future.
	}, testcase.Flaky(42))
}

func ExampleNewT() {
	variable := testcase.Var[int]{ID: "variable", Init: func(t *testcase.T) int {
		return t.Random.Int()
	}}

	// flat test case with test runtime variable caching
	var tb testing.TB
	t := testcase.NewT(tb, testcase.NewSpec(tb))
	value1 := variable.Get(t)
	value2 := variable.Get(t)
	t.Logf(`test case variable caching works even in flattened tests: v1 == v2 -> %v`, value1 == value2)
}

func Example_faultInject() {
	defer faultinject.Enable()()

	ctx := context.Background()

	// all fault field is optional.
	// in case left as zero value,
	// it will match every caller context,
	// and returns on the first .Err() / .Value(faultinject.Fault{})
	ctx = faultinject.Inject(ctx, faultinject.CallerFault{
		Package:  "targetpkg",
		Receiver: "*myreceiver",
		Function: "myfunction",
	}, errors.New("boom"))

	// from and after call stack reached: targetpkg.(*myreceiver).myfunction
	if err := ctx.Err(); err != nil {
		fmt.Println(err) // in the position defined by the Fault, it will yield an error
	}
}

func Example_assertEventually() {
	waiter := assert.Waiter{
		WaitDuration: time.Millisecond,
		Timeout:      time.Second,
	}
	w := assert.Eventually{RetryStrategy: waiter}

	var t *testing.T
	// will attempt to wait until assertion block passes without a failing testCase result.
	// The maximum time it is willing to wait is equal to the wait timeout duration.
	// If the wait timeout reached, and there was no passing assertion run,
	// the last failed assertion history is replied to the received testing.TB
	//   In this case the failure would be replied to the *testing.T.
	w.Assert(t, func(it assert.It) {
		if rand.Intn(1) == 0 {
			it.Fatal(`boom`)
		}
	})
}

func Example_assertEventuallyAsContextOption() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test(`flaky`, func(t *testcase.T) {
		// flaky test content here
	}, testcase.Flaky(assert.RetryCount(42)))
}

func Example_assertEventuallyCount() {
	_ = assert.Eventually{RetryStrategy: assert.RetryCount(42)}
}

func Example_assertEventuallyByTimeout() {
	r := assert.Eventually{RetryStrategy: assert.Waiter{
		WaitDuration: time.Millisecond,
		Timeout:      time.Second,
	}}

	var t *testing.T
	r.Assert(t, func(it assert.It) {
		if rand.Intn(1) == 0 {
			it.Fatal(`boom`)
		}
	})
}

func Example_assertEventuallyByCount() {
	r := assert.Eventually{RetryStrategy: assert.RetryCount(42)}

	var t *testing.T
	r.Assert(t, func(it assert.It) {
		if rand.Intn(1) == 0 {
			it.Fatal(`boom`)
		}
	})
}

func Example_assertEventuallyByCustomRetryStrategy() {
	// this approach ideal if you need to deal with asynchronous systems
	// where you know that if a workflow process ended already,
	// there is no point in retrying anymore the assertion.

	while := func(isFailed func() bool) {
		for isFailed() {
			// just retry while assertion is failed
			// could be that assertion will be failed forever.
			// Make sure the assertion is not stuck in a infinite loop.
		}
	}

	r := assert.Eventually{RetryStrategy: assert.RetryStrategyFunc(while)}

	var t *testing.T
	r.Assert(t, func(it assert.It) {
		if rand.Intn(1) == 0 {
			it.Fatal(`boom`)
		}
	})
}

func ExampleSetEnv() {
	var tb testing.TB
	testcase.SetEnv(tb, `MY_KEY`, `myvalue`)
	// env will be restored after the test
}

func ExampleUnsetEnv() {
	var tb testing.TB
	testcase.UnsetEnv(tb, `MY_KEY`)
	// env will be restored after the test
}

func ExampleAppend() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	list := testcase.Let(s, func(t *testcase.T) interface{} {
		return []int{}
	})

	s.Before(func(t *testcase.T) {
		t.Log(`some context where a value is expected in the testcase.Var[[]T] variable`)
		testcase.Append(t, list, 42)
	})

	s.Test(``, func(t *testcase.T) {
		t.Log(`list will include the appended value`)
		list.Get(t) // []int{42}
	})
}

func ExampleSpec_whenProjectUseSharedSpecificationHelpers() {
	var t *testing.T
	s := testcase.NewSpec(t)

	myType := func() *mydomain.MyUseCase { return &mydomain.MyUseCase{} }

	s.Describe(`#MyFunc`, func(s *testcase.Spec) {
		var (
			something = spechelper.GivenWeHaveSomething(s)
			// .. other givens
		)
		act := func(t *testcase.T) {
			myType().MyFuncThatNeedsSomething(something.Get(t))
		}

		s.Then(`test case described here`, func(t *testcase.T) {
			act(t)
		})
	})
}

func ExampleVar_Super() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	v := testcase.Let[int](s, func(t *testcase.T) int {
		return 32
	})

	s.Context("some sub context", func(s *testcase.Spec) {
		v.Let(s, func(t *testcase.T) int {
			return v.Super(t) + 10 // where super == 32 from the parent context
		})

		s.Test("the result of the V", func(t *testcase.T) {
			t.Must.Equal(42, v.Get(t))
		})
	})
}

func ExampleTableTest_classicInlined() {
	myFunc := func(in int) string {
		if in != 42 {
			return "Not the Answer"
		}
		return "The Answer"
	}
	var tb testing.TB
	type TTCase struct {
		In       int
		Expected string
	}
	testcase.TableTest(tb, map[string]TTCase{
		"when A": {
			In:       42,
			Expected: "The Answer",
		},
		"when B": {
			In:       24,
			Expected: "Not the Answer",
		},
		"when C": {
			In:       128,
			Expected: "Not the Answer",
		},
	}, func(t *testcase.T, tc TTCase) {
		got := myFunc(tc.In)
		t.Must.Equal(tc.Expected, got)
	})
}

func ExampleTableTest_classicStructured() {
	var tb testing.TB
	myFunc := func(in int) string {
		if in == 42 {
			return "The Answer"
		}
		return "Not the answer"
	}
	type Case struct {
		Input    int
		Expected string
	}
	arrangements := map[string]Case{
		"when the input is correct": {
			Input:    42,
			Expected: "The Answer",
		},
		"when something else 1": {
			Input:    24,
			Expected: "Not the answer",
		},
		"when someting else 2": {
			Input:    128,
			Expected: "Not the answer",
		},
	}
	act := func(t *testcase.T, tc Case) {
		got := myFunc(tc.Input)
		t.Must.Equal(tc.Expected, got)
	}
	testcase.TableTest(tb, arrangements, act)
}

func ExampleTableTest_withSpecBlock() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	var (
		in = testcase.Let[int](s, nil)
	)
	act := func(t *testcase.T) {
		// my act that use in
		_ = in.Get(t)
	}

	testcase.TableTest(s, map[string]func(s *testcase.Spec){
		"when 42": func(s *testcase.Spec) {
			in.LetValue(s, 42)
		},
		"whe 24": func(s *testcase.Spec) {
			in.LetValue(s, 42)
		},
	}, act)
}

func ExampleTableTest_withTestBlock() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	var (
		in = testcase.Let[int](s, nil)
	)
	act := func(t *testcase.T) {
		// my act that use in
		_ = in.Get(t)
	}

	testcase.TableTest(s, map[string]func(t *testcase.T){
		"when 42": func(t *testcase.T) {
			in.Set(t, 42)
		},
		"whe 24": func(t *testcase.T) {
			in.Set(t, 24)
		},
	}, act)
}

func ExampleT_SetEnv() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test("", func(t *testcase.T) {
		t.SetEnv("key", "value")
	})
}

func ExampleT_UnsetEnv() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test("", func(t *testcase.T) {
		t.UnsetEnv("key")
	})
}
