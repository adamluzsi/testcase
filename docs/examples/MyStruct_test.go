package examples_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/docs/examples"
)

var myStruct = testcase.Var{
	Name: `example MyStruct`,
	Init: func(t *testcase.T) interface{} {
		return examples.MyStruct{}
	},
}

func myStructGet(t *testcase.T) examples.MyStruct {
	return myStruct.Get(t).(examples.MyStruct)
}

func TestMyStruct(t *testing.T) {
	s := testcase.NewSpec(t)
	s.NoSideEffect()

	// define shared variables and hooks here
	// ...
	myStruct.Let(s, nil)

	s.Describe(`Say`, SpecMyStruct_Say)
	s.Describe(`Foo`, SpecMyStruct_Foo)
	s.Describe(`Bar`, SpecMyStruct_Bar)
	s.Describe(`Baz`, SpecMyStruct_Baz)
	// other specification sub contexts
}

func SpecMyStruct_Say(s *testcase.Spec) {
	var subject = func(t *testcase.T) string {
		return myStructGet(t).Say()
	}

	s.Then(`it will say a famous quote`, func(t *testcase.T) {
		assert.Must(t).Equal(`Hello, World!`, subject(t))
	})
}

func SpecMyStruct_Foo(s *testcase.Spec) {
	var subject = func(t *testcase.T) string {
		return myStructGet(t).Foo()
	}

	s.Then(`it will return with Foo`, func(t *testcase.T) {
		assert.Must(t).Equal(`Foo`, subject(t))
	})
}

func SpecMyStruct_Bar(s *testcase.Spec) {
	var subject = func(t *testcase.T) string {
		return myStructGet(t).Bar()
	}

	s.Then(`it will return with Bar`, func(t *testcase.T) {
		assert.Must(t).Equal(`Bar`, subject(t))
	})
}

func SpecMyStruct_Baz(s *testcase.Spec) {
	var subject = func(t *testcase.T) string {
		return myStructGet(t).Baz()
	}

	s.Then(`it will return with Baz`, func(t *testcase.T) {
		assert.Must(t).Equal(`Baz`, subject(t))
	})
}
