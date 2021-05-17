package random_test

import (
	"github.com/adamluzsi/testcase/random"
	"math/rand"
	"time"
)

func ExampleRandom_mathRand() {
	seed := time.Now().UnixNano()
	source := rand.NewSource(seed)
	random.NewRandom(source)
}

func ExampleRandom_cryptoSeed() {
	random.NewRandom(random.CryptoSeed{})
}
