package let_test

import (
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/sandbox"
)

func TestAs(t *testing.T) {
	t.Run("primitive type", func(t *testing.T) {
		type MyString string

		s := testcase.NewSpec(t)
		v1 := let.String(s)
		v2 := let.As[MyString](v1)

		s.Test("", func(t *testcase.T) {
			t.Must.Equal(MyString(v1.Get(t)), v2.Get(t))
		})
	})

	t.Run("interface type", func(t *testing.T) {
		type TimeAfterer interface {
			After(u time.Time) bool
		}

		s := testcase.NewSpec(t)
		v1 := let.Time(s)
		v2 := let.As[TimeAfterer](v1)

		s.Test("", func(t *testcase.T) {
			t.Must.Equal(TimeAfterer(v1.Get(t)), v2.Get(t))
		})
	})

	t.Run("panics on incorrect conversation", func(t *testing.T) {
		ro := sandbox.Run(func() {
			s := testcase.NewSpec(t)
			v1 := let.Time(s)
			_ = let.As[string](v1)
		})
		assert.False(t, ro.OK)
		assert.False(t, ro.Goexit)
		assert.NotNil(t, ro.PanicValue)
	})
}
