package random_test

import (
	"github.com/adamluzsi/testcase/random"
	"math/rand"
	"time"
)

func ExampleRandom_mathRand() {
	seed := time.Now().UnixNano()
	source := rand.NewSource(seed)
	random.New(source)
}

func ExampleRandom_cryptoSeed() {
	random.New(random.CryptoSeed{})
}
