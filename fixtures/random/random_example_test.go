package random_test

import (
	"math/rand"
	"time"

	"github.com/adamluzsi/testcase/fixtures/random"
)

func ExampleNewRandom() {
	_ = random.NewRandom(rand.NewSource(time.Now().Unix()))
}

func ExampleRandom_Bool() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	_ = rnd.Bool() // returns a random bool
}

func ExampleRandom_ElementFromSlice() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	// returns a random element from the given slice
	_ = rnd.ElementFromSlice([]string{`foo`, `bar`, `baz`}).(string)
}

func ExampleRandom_Float32() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	_ = rnd.Float32()
}

func ExampleRandom_Float64() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	_ = rnd.Float64()
}

func ExampleRandom_Int() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	_ = rnd.Int()
}

func ExampleRandom_IntBetween() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	_ = rnd.IntBetween(24, 42)
}

func ExampleRandom_IntN() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	_ = rnd.IntN(42)
}

func ExampleRandom_KeyFromMap() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	_ = rnd.KeyFromMap(map[string]struct{}{
		`foo`: {},
		`bar`: {},
		`baz`: {},
	}).(string)
}

func ExampleRandom_String() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	_ = rnd.String()
}

func ExampleRandom_StringN() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	_ = rnd.StringN(42)
}

func ExampleRandom_Time() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	_ = rnd.Time()
}

func ExampleRandom_TimeBetween() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	_ = rnd.TimeBetween(time.Now(), time.Now().Add(time.Hour))
}
