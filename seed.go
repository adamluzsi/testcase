package testcase

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/adamluzsi/testcase/internal"
)

func makeSeed() (int64, error) {
	rawSeed, injectedRandomSeedIsSet := os.LookupEnv(EnvKeySeed)
	if !injectedRandomSeedIsSet {
		return time.Now().UnixNano(), nil
	}
	seed, err := strconv.ParseInt(rawSeed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s has invalid seed integer value: %s", EnvKeySeed, rawSeed)
	}
	return seed, nil
}

func seedForSpec(tb testing.TB) (_seed int64) {
	tb.Helper()
	tb.Cleanup(func() {
		tb.Helper()
		if tb.Failed() {
			// Help developers to know the seed of the failed test execution.
			internal.Log(tb, fmt.Sprintf(`%s=%d`, EnvKeySeed, _seed))
		}
	})
	seed, err := makeSeed()
	if err != nil {
		tb.Fatal(err.Error())
	}
	return seed
}
