package random_test

import (
	"math/rand"
	"testing"

	"github.com/adamluzsi/testcase/random"

	"github.com/adamluzsi/testcase"
)

var _ rand.Source64 = random.CryptoSeed{}

func TestCryptoSeed(t *testing.T) {
	s := testcase.NewSpec(t)

	var seed = func(t *testcase.T) random.CryptoSeed {
		return random.CryptoSeed{}
	}

	s.Describe(`usage with Random`, func(s *testcase.Spec) {
		s.Let(`randomizer`, func(t *testcase.T) interface{} {
			return random.NewRandom(seed(t))
		})

		SpecRandomizerMethods(s)
	})
}
