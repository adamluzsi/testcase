package testcase

import (
	"fmt"
	"github.com/adamluzsi/testcase/internal"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

// EnvKeyOrderSeed is the environment variable key that will be checked for a pseudo random seed,
// which will be used to randomize the order of executions between testCase cases.
const EnvKeyOrderSeed = `TESTCASE_ORDER_SEED`

// EnvKeyOrderMod is the environment variable key that will be checked for testCase determine
// what order of execution should be used between testCase cases in a testing group.
// The default sorting behavior is pseudo random based on an the seed.
//
// Mods:
// - defined: execute testCase in the order which they are being defined
// - random: pseudo random based ordering between tests.
const EnvKeyOrderMod = `TESTCASE_ORDER_MOD`

//-------------------------------------------------- Env Var Helpers -------------------------------------------------//

// SetEnv will set the os environment variable for the current program to a given value,
// and prepares a cleanup function to restore the original state of the environment variable.
//
// Spec using this helper should be flagged with Spec.HasSideEffect or Spec.Sequential.
func SetEnv(tb testing.TB, key, value string) {
	cleanupEnv(tb, key)

	if err := os.Setenv(key, value); err != nil {
		tb.Fatal(err)
	}
}

// UnsetEnv will unset the os environment variable value for the current program,
// and prepares a cleanup function to restore the original state of the environment variable.
//
// Spec using this helper should be flagged with Spec.HasSideEffect or Spec.Sequential.
func UnsetEnv(tb testing.TB, key string) {
	cleanupEnv(tb, key)

	if err := os.Unsetenv(key); err != nil {
		tb.Fatal(err)
	}
}

func cleanupEnv(tb testing.TB, key string) {
	var restore func() error
	if originalValue, ok := os.LookupEnv(key); ok {
		restore = func() error { return os.Setenv(key, originalValue) }
	} else {
		restore = func() error { return os.Unsetenv(key) }
	}
	tb.Cleanup(func() {
		if err := restore(); err != nil {
			tb.Error(err)
		}
	})
}

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
	if !internal.CacheEnabled {
		globalOrderModInit = sync.Once{}
	}

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
			panic(fmt.Sprintf(`unknown testCase ordering/arrange mod: %s`, mod))
		}
	})

	tb.Logf(` Test Execution Seed: %s`, globalOrderMod)
	return globalOrderMod
}
