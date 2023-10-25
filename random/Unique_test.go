package random_test

import (
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/clock/timecop"
	"go.llib.dev/testcase/random"
	"go.llib.dev/testcase/sandbox"
	"testing"
)

func ExampleUnique() {
	// useful when you need random values which are not equal
	rnd := random.New(random.CryptoSeed{})
	v1 := rnd.Int()
	v2 := random.Unique(rnd.Int, v1)
	v3 := random.Unique(rnd.Int, v1, v2)

	var tb testing.TB
	assert.NotEqual(tb, v1, v3)
	assert.NotEqual(tb, v2, v3)
}

func TestUnique(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	t.Run("no exclude list given", func(t *testing.T) {
		v := random.Unique(rnd.Int)
		assert.NotEmpty(t, v)
	})
	t.Run("exclude list has a value", func(t *testing.T) {
		rnd.Repeat(128, 256, func() {
			v1 := rnd.IntBetween(1, 3)
			v2 := random.Unique(func() int {
				return rnd.IntBetween(1, 3)
			}, v1)
			assert.NotEqual(t, v1, v2)
		})
	})
	t.Run("exclude list has multiple values", func(t *testing.T) {
		rnd.Repeat(128, 256, func() {
			v1 := 0
			v2 := 1
			v3 := 2
			got := random.Unique(func() int {
				return rnd.IntBetween(0, 3)
			}, v1, v2, v3)
			assert.NotEqual(t, got, v1)
			assert.NotEqual(t, got, v2)
			assert.NotEqual(t, got, v3)
		})
	})
	t.Run("If the function takes too long to find a valid value, it will trigger a panic once a set time limit is reached", func(t *testing.T) {
		timecop.SetSpeed(t, timecop.BlazingFast)
		var ran bool
		out := sandbox.Run(func() {
			random.Unique(func() int {
				ran = true
				return 0
			}, 0)
		})
		assert.True(t, ran)
		assert.False(t, out.OK)
		assert.NotEmpty(t, out.PanicValue)
	})
}
