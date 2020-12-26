package testcase

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

// EnvKeyOrderSeed is the environment variable key that will be checked for a pseudo random seed,
// which will be used to randomize the order of executions between test cases.
const EnvKeyOrderSeed = `TESTCASE_ORDER_SEED`

// EnvKeyOrderMod is the environment variable key that will be checked for test determine
// what order of execution should be used between test cases in a testing group.
// The default sorting behavior is pseudo random based on an the seed.
//
// Mods:
// - defined: execute test in the order which they are being defined
// - random: pseudo random based ordering between tests.
const EnvKeyOrderMod = `TESTCASE_ORDER_MOD`

//---------------------------------------------- Test Sorter Random Seed ---------------------------------------------//

var (
	globalRandomOrderSeed     int64
	globalRandomOrderSeedInit sync.Once
)

func getGlobalRandomOrderSeed(tb testing.TB) int64 {
	globalRandomOrderSeedInit.Do(func() {
		rawSeed, ok := os.LookupEnv(EnvKeyOrderSeed)
		if !ok {
			globalRandomOrderSeed = time.Now().Unix()
			return
		}

		seed, err := strconv.ParseInt(rawSeed, 10, 64)
		require.Nil(tb, err)
		globalRandomOrderSeed = seed
	})

	tb.Logf(`Test Random Order Seed: %d`, globalRandomOrderSeed)
	return globalRandomOrderSeed
}

//---------------------------------------------- Global Test Sorter Mod ----------------------------------------------//

var (
	globalOrderMod     testOrderingMod
	globalOrderModInit sync.Once
)

func getGlobalOrderMod(tb testing.TB) testOrderingMod {
	globalOrderModInit.Do(func() {
		mod, ok := os.LookupEnv(EnvKeyOrderMod)
		if !ok {
			globalOrderMod = OrderingAsRandom
			return
		}

		switch testOrderingMod(mod) {
		case OrderingAsDefined:
			globalOrderMod = OrderingAsDefined
		case OrderingAsRandom:
			globalOrderMod = OrderingAsRandom
		default:
			panic(fmt.Sprintf(`unknown test ordering/arrange mod: %s`, mod))
		}
	})

	tb.Logf(` Test Execution Seed: %s`, globalOrderMod)
	return globalOrderMod
}
