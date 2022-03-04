package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleT_Defer_withArgs() {
	var t *testing.T
	s := testcase.NewSpec(t)

	const something = `the ExampleDeferTeardownWithArgs value`

	testcase.Let(s, func(t *testcase.T) interface{} {
		ptr := &ExampleDeferTeardownWithArgs{}
		// T#Defer arguments copied upon pass by value
		// and then passed to the function during the execution of the deferred function call.
		//
		// This is ideal for situations where you need to guarantee that a value cannot be muta
		t.Defer(ptr.SomeTeardownWithArg, `Hello, World!`)
		return ptr
	})

	s.Test(`a simple test case`, func(t *testcase.T) {
		entity := t.I(something).(*ExampleDeferTeardownWithArgs)

		entity.DoSomething()
	})
}

type ExampleDeferTeardownWithArgs struct{}

func (*ExampleDeferTeardownWithArgs) SomeTeardownWithArg(arg string) {}

func (*ExampleDeferTeardownWithArgs) DoSomething() {}
