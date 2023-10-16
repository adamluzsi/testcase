package let_test

import (
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/random"
)

func TestWith(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	t.Run("func() V", func(t *testing.T) {
		s := testcase.NewSpec(t)
		n := rnd.Int()
		v := let.With[int](s, func() int {
			return n
		})
		s.Test("", func(t *testcase.T) {
			t.Must.Equal(n, v.Get(t))
		})
	})
	t.Run("func(testing.TB) V", func(t *testing.T) {
		s := testcase.NewSpec(t)
		n := rnd.String()
		v := let.With[string](s, func(testing.TB) string {
			return n
		})
		s.Test("", func(t *testcase.T) {
			t.Must.Equal(n, v.Get(t))
		})
	})
	t.Run("func(*testcase.T) V", func(t *testing.T) {
		s := testcase.NewSpec(t)
		n := let.UUID(s)
		v := let.With[string](s, func(t *testcase.T) string {
			return n.Get(t)
		})
		s.Test("", func(t *testcase.T) {
			t.Must.Equal(n.Get(t), v.Get(t))
		})
	})
}
