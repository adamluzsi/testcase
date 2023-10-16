package random_test

import (
	"math/rand"
	"testing"
	"time"

	"go.llib.dev/testcase/pp"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/random"
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

func ExampleRandom_SliceElement() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	// returns a random element from the given slice
	_ = rnd.SliceElement([]string{`foo`, `bar`, `baz`}).(string)
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

func ExampleRandom_String() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.String()
}

func ExampleRandom_StringN() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))

	_ = rnd.StringN(42)
}

func ExampleRandom_StringNWithCharset() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))
	rnd.StringNWithCharset(42, random.Charset())
	rnd.StringNWithCharset(42, random.CharsetASCII())
	rnd.StringNWithCharset(42, random.CharsetAlpha())
	rnd.StringNWithCharset(42, "ABC")
}

func ExampleRandom_StringNC() {
	rnd := random.New(rand.NewSource(time.Now().Unix()))
	rnd.StringNC(42, random.Charset())
	rnd.StringNC(42, random.CharsetASCII())
	rnd.StringNC(42, random.CharsetAlpha())
	rnd.StringNC(42, "ABC")
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

func ExampleMakeSlice() {
	rnd := random.New(random.CryptoSeed{})

	pp.PP(random.Slice[int](3, rnd.Int)) // []int slice with 3 values
}

func ExampleMakeMap() {
	rnd := random.New(random.CryptoSeed{})

	pp.PP(random.Map[string, int](3, random.KV(rnd.String, rnd.Int))) // map[string]int slice with 3 key-value pairs
}

func ExampleRandom_Repeat() {
	rnd := random.New(random.CryptoSeed{})

	n := rnd.Repeat(1, 3, func() {
		// will be called repeatedly between 1 and 3 times.
	})

	_ = n // is the number of times, the function block was repeated.
}
