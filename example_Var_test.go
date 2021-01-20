package testcase_test

import (
	"github.com/adamluzsi/testcase"
	"testing"
)

func ExampleVar() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		resource = testcase.Var{Name: `resource`}
		myType   = s.Let(`myType`, func(t *testcase.T) interface{} {
			return &MyType{MyResource: resource.Get(t).(RoleInterface)}
		})
	)

	s.Describe(`#MyFunction`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			// after GO2 this will be replaced with concrete Types instead of interface{}
			myType.Get(t).(*MyType).MyFunc()
		}

		s.When(`resource is xy`, func(s *testcase.Spec) {
			resource.Let(s, func(t *testcase.T) interface{} {
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

	value := s.Let(`some value`, func(t *testcase.T) interface{} {
		return 42
	})

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t).(int) // -> 42
	})
}

func ExampleVar_Set() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := s.Let(`some value`, func(t *testcase.T) interface{} {
		return 42
	})

	s.Before(func(t *testcase.T) {
		value.Set(t, 24)
	})

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t).(int) // -> 24
	})
}

func ExampleVar_Let() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Var{
		Name: `the variable group`,
		Init: func(t *testcase.T) interface{} {
			return 42
		},
	}

	value.Let(s, nil)

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t).(int) // -> 42
	})
}

func ExampleVar_Let_valueDefinedAtTestingContextScope() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Var{Name: `the variable group`}

	value.Let(s, func(t *testcase.T) interface{} {
		return 42
	})

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t).(int) // -> 42
	})
}

func ExampleVar_LetValue() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Var{Name: `the variable group`}

	value.LetValue(s, 42)

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t).(int) // -> 42
	})
}

func ExampleVar_EagerLoading() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := s.Let(`some value`, func(t *testcase.T) interface{} {
		return 42
	})

	// will be loaded early on, before the testCase case block reached.
	// This can be useful when you want to have variables,
	// that also must be present in some sort of attached resource,
	// and as part of the constructor, you want to save it.
	// So when the testCase block is reached, the entity is already present in the resource.
	value.EagerLoading(s)

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t).(int) // -> 42
		// value returned from cache instead of triggering first time initialization.
	})
}

func ExampleVar_Let_eagerLoading() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Var{Name: `value`}

	value.Let(s, func(t *testcase.T) interface{} {
		return 42
	}).EagerLoading(s)

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t).(int) // -> 42
		// value returned from cache instead of triggering first time initialization.
	})
}

func ExampleVar_LetValue_eagerLoading() {
	var t *testing.T
	s := testcase.NewSpec(t)

	value := testcase.Var{Name: `value`}
	value.LetValue(s, 42).EagerLoading(s)

	s.Test(`some testCase`, func(t *testcase.T) {
		_ = value.Get(t).(int) // -> 42
		// value returned from cache instead of triggering first time initialization.
	})
}
