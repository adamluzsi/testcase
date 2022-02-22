package testcase_test

import (
	"database/sql"
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleVar() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		resource = testcase.Var[MyResourceSupplier]{Name: `resource`}
		myType   = testcase.Let(s, `myType`, func(t *testcase.T) *MyType {
			return &MyType{MyResource: resource.Get(t)}
		})
	)

	s.Describe(`#MyFunction`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			// after GO2 this will be replaced with concrete Types instead of interface{}
			myType.Get(t).MyFunc()
		}

		s.When(`resource is xy`, func(s *testcase.Spec) {
			resource.Let(s, func(t *testcase.T) MyResourceSupplier {
				return MyResourceSupplier{}
			})

			s.Then(`do some testCase`, func(t *testcase.T) {
				subject(t) // act
				// assertions here.
			})
		})

		// ...
		// other cases with resource xy state change
	})
}

func ExampleVar_Get() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Let(s, `some value`, func(t *testcase.T) interface{} {
		return 42
	})

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t) // -> 42
	})
}

func ExampleVar_Set() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Let(s, `some value`, func(t *testcase.T) interface{} {
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
		Name: `the variable group`,
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

	value := testcase.Var[int]{Name: `the variable group`}

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

	value := testcase.Var[int]{Name: `the variable group`}

	value.LetValue(s, 42)

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t) // -> 42
	})
}

func ExampleVar_EagerLoading() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Let(s, `some value`, func(t *testcase.T) interface{} {
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

	value := testcase.Var[int]{Name: `value`}

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

	value := testcase.Var[int]{Name: `value`}
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
		Name: `value`,
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
		Name: `db`,
		Init: func(t *testcase.T) *sql.DB {
			db, err := sql.Open(`driver`, `dataSourceName`)
			if err != nil {
				t.Fatal(err.Error())
			}
			return db
		},
		OnLet: func(s *testcase.Spec) {
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
	v := testcase.Var[int]{Name: "myvar", Init: func(t *testcase.T) int { return 42 }}
	v.Bind(s)
	s.Test(``, func(t *testcase.T) {
		_ = v.Get(t) // -> 42
	})
}

func ExampleVar_before() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	v := testcase.Var[int]{
		Name: "myvar",
		Init: func(t *testcase.T) int { return 42 },
		Before: func(t *testcase.T) {
			t.Log(`I'm from the Var.Before block`)
		},
	}
	s.Test(``, func(t *testcase.T) {
		_ = v.Get(t)
		// log: I'm from the Var.Before block
		// -> 42
	})
}
