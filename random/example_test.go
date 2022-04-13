package random_test

import (
	"math/rand"
	"time"

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
