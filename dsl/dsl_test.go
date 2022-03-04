package dsl_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/testcase/dsl"
)

func Test(t *testing.T) {
	Spec(t).Describe(`smoke testing of testcase DSL`, func(s *testcase.Spec) {
		num := Let[int](s, func(t *testcase.T) int {
			return t.Random.Int() + 1
		})
		str := LetValue[string](s, "42")

		s.Test(``, func(t *testcase.T) {
			Should(t).Equal("42", str.Get(t))
			Must(t).NotEqual(0, num.Get(t))
		})
	})
}
