package random_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/random"
)

func ExampleRandom_mathRand() {
	seed := time.Now().UnixNano()
	source := rand.NewSource(seed)
	random.New(source)
}

func ExampleRandom_cryptoSeed() {
	random.New(random.CryptoSeed{})
}

func ExampleRandom_Make() {
	rnd := random.New(random.CryptoSeed{})

	type ExampleStruct struct {
		Foo string
		Bar int
		Baz *int
	}

	_ = rnd.Make(&ExampleStruct{}).(*ExampleStruct) // returns a populated struct
}

func ExampleNew() {
	_ = random.New(rand.NewSource(time.Now().Unix()))
}

func ExampleRandom_Bool() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.Bool() // returns a random bool
}

func ExampleRandom_ElementFromSlice() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	// returns a random element from the given slice
	_ = rnd.ElementFromSlice([]string{`foo`, `bar`, `baz`}).(string)
}

func ExampleRandom_Float32() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.Float32()
}

func ExampleRandom_Float64() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.Float64()
}

func ExampleRandom_Int() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.Int()
}

func ExampleRandom_IntBetween() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.IntBetween(24, 42)
}

func ExampleRandom_IntN() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.IntN(42)
}

func ExampleRandom_KeyFromMap() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.KeyFromMap(map[string]struct{}{
		`foo`: {},
		`bar`: {},
		`baz`: {},
	}).(string)
}

func ExampleRandom_String() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.String()
}

func ExampleRandom_StringN() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.StringN(42)
}

func ExampleRandom_Time() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.Time()
}

func ExampleRandom_TimeBetween() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.TimeBetween(time.Now(), time.Now().Add(time.Hour))
}

func ExampleRandom_TimeN() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))
	var (
		years  = 0
		months = 4
		days   = 2
	)
	_ = rnd.TimeN(time.Now(), years, months, days)
}

func ExampleRandom_Read() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	p := make([]byte, 42)
	n, err := rnd.Read(p)

	_, _ = n, err
}

func ExampleRandom_Error() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	err := rnd.Error()
	_ = err
}

func TestExampleRandomError(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Test("", func(t *testcase.T) {
		err := t.Random.Error()
		t.Log(err.Error())
	})
}
