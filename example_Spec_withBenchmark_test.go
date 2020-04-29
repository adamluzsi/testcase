package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_withBenchmark() {
	var b *testing.B
	s := testcase.NewSpec(b)

	myType := func(t *testcase.T) *MyType {
		return &MyType{Field1: `Hello, World!`}
	}

	s.When(`something`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Log(`setup`)
		})

		s.Then(`this benchmark block will be executed by *testing.B.N times`, func(t *testcase.T) {
			myType(t).IsLower()
		})
	})
}
