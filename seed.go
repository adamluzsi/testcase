package testcase

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"go.llib.dev/testcase/random"

	"go.llib.dev/testcase/internal"
	"go.llib.dev/testcase/internal/environ"
)

func makeSeed() (int64, error) {
	rawSeed, injectedRandomSeedIsSet := os.LookupEnv(environ.KeySeed)
	if !injectedRandomSeedIsSet {
		salt := rand.New(random.CryptoSeed{}).Int63()
		base := time.Now().UnixNano()
		return base + salt, nil
	}
	seed, err := strconv.ParseInt(rawSeed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s has invalid seed integer value: %s", environ.KeySeed, rawSeed)
	}
	return seed, nil
}

func seedForSpec(tb testing.TB) (_seed int64) {
	helper(tb).Helper()
	if isValidTestingTB(tb) {
		tb.Cleanup(func() {
			tb.Helper()
			if tb.Failed() {
				// Help developers to know the seed of the failed test execution.
				internal.Log(tb, fmt.Sprintf(`%s=%d`, environ.KeySeed, _seed))
			}
		})
	}
	seed, err := makeSeed()
	if err != nil {
		tb.Fatal(err.Error())
	}
	return seed
}

func isValidTestingTB(tb testing.TB) bool {
	if tb == nil {
		return false
	}
	_, ok := tb.(internal.NullTB)
	return !ok
}
