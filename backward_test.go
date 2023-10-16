package testcase

import (
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/random"
)

func TestRetryCount(t *testing.T) {
	_ = RetryCount(42)
}

func TestSpec_Let_andLetValue_backwardCompatibility(t *testing.T) {
	s := NewSpec(t)

	rnd := random.New(random.CryptoSeed{})
	r1 := rnd.Int()
	r2 := rnd.Int()

	v1 := s.Let(`answer`, func(t *T) interface{} { return r1 })
	v2 := s.LetValue(`count`, r2)

	s.Test(``, func(t *T) {
		t.Must.Equal(r1, v1.Get(t))
		t.Must.Equal(r2, v2.Get(t))
	})
}

func TestSpec_LetValue_returnsVar(t *testing.T) {
	s := NewSpec(t)

	counter := s.LetValue(`counter`, 0)

	s.Test(``, func(t *T) {
		assert.Must(t).Equal(0, counter.Get(t))
		counter.Set(t, 1)
		assert.Must(t).Equal(1, counter.Get(t))
		counter.Set(t, 2)
		assert.Must(t).Equal(2, counter.Get(t))
	})
}
