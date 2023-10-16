package random_test

import (
	"math/rand"
	"testing"
	"time"

	"go.llib.dev/testcase/random"

	"go.llib.dev/testcase"
)

var _ rand.Source64 = random.CryptoSeed{}

func TestCryptoSeed(t *testing.T) {
	s := testcase.NewSpec(t)

	var seed = func(t *testcase.T) random.CryptoSeed {
		return random.CryptoSeed{}
	}

	s.Describe(`usage with Random`, func(s *testcase.Spec) {
		randomizer := testcase.Let(s, func(t *testcase.T) *random.Random {
			return random.New(seed(t))
		})

		SpecRandomMethods(s, randomizer)
	}, testcase.Flaky(5*time.Second))
}
