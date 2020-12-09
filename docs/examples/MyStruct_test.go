package examples_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/docs/examples"
)

func TestMyStruct(t *testing.T) {
	s := testcase.NewSpec(t)
	s.NoSideEffect()

	s.Let(`my-struct`, func(t *testcase.T) interface{} {
		return examples.MyStruct{}
	})

	// define shared variables and hooks here
	// ...

	s.Describe(`Say`, SpecMyStruct_Say)
	s.Describe(`Foo`, SpecMyStruct_Foo)
	// other specification sub contexts
}

func SpecMyStruct_Say(s *testcase.Spec) {
	var subject = func(t *testcase.T) string {
		return t.I(`my-struct`).(examples.MyStruct).Say()
	}

	s.Then(`it will say a famous quote`, func(t *testcase.T) {
		require.Equal(t, `Hello, World!`, subject(t))
	})
}

func SpecMyStruct_Foo(s *testcase.Spec) {
	var subject = func(t *testcase.T) string {
		return t.I(`my-struct`).(examples.MyStruct).Foo()
	}

	s.Then(`it will say a famous quote`, func(t *testcase.T) {
		require.Equal(t, `Foo`, subject(t))
	})
}
