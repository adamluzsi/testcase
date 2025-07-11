package random_test

import (
	"net/netip"
	"testing"
	"time"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/clock/timecop"
	"go.llib.dev/testcase/internal/proxy"
	"go.llib.dev/testcase/random"
	"go.llib.dev/testcase/sandbox"
)

var rnd = random.New(random.CryptoSeed{})

func ExampleUnique() {
	// useful when you need random values which are not equal
	v1 := rnd.Int()
	v2 := random.Unique(rnd.Int, v1)
	v3 := random.Unique(rnd.Int, v1, v2)

	var tb testing.TB
	assert.NotEqual(tb, v1, v3)
	assert.NotEqual(tb, v2, v3)
}

func TestUnique(t *testing.T) {
	now := time.Now()
	var i time.Duration
	proxy.StubTimeNow(t, func() time.Time {
		i++
		return now.Add(time.Millisecond * i)
	})

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
	t.Run("creating an item takes a lot of time then instead of time based retry, we make at least 5 attempts", func(t *testing.T) {
		now := time.Now()
		timecop.Travel(t, now, timecop.Freeze)
		out := sandbox.Run(func() {
			var i int
			random.Unique(func() int {
				timecop.Travel(t, 10*time.Second)
				i++
				if 5 <= i {
					return i
				}
				return 0
			}, 0)
		})
		assert.True(t, out.OK)
	})
	t.Run("if the unique's make function is fast enough, then more than 5 tries will be made, as long it can fit within the deadline", func(t *testing.T) {
		timecop.SetSpeed(t, 1000 /* times */)
		var n int
		out := sandbox.Run(func() {
			random.Unique(func() int {
				n++
				return 0
			}, 0)
		})
		assert.False(t, out.OK)
		assert.True(t, 6 < n) // probably it runs at least 20000 times, so it should be definetly bigger than 6
	})
	t.Run("comparison of struct values without exported fields", func(t *testing.T) {
		p1, err := netip.ParsePrefix("10.0.0.0/24")
		assert.NoError(t, err)

		p2, err := netip.ParsePrefix("10.0.0.1/24")
		assert.NoError(t, err)

		assert.NotPanic(t, func() {
			random.Unique(func() netip.Prefix {
				return p1
			}, p2)
		})

		assert.Panic(t, func() {
			random.Unique(func() netip.Prefix {
				return p1
			}, p1)
		})
	})
}
