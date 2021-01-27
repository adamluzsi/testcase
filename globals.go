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

	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Logf(`Test Random Order Seed: %d`, globalRandomOrderSeed)
		}
	})
	return globalRandomOrderSeed
}

//---------------------------------------------- Global Test Sorter Mod ----------------------------------------------//

var (
	globalOrderMod     testOrderingMod
	globalOrderModInit sync.Once
)

func getGlobalOrderMod(tb testing.TB) testOrderingMod {
	if !internal.CacheEnabled {
		return getOrderModFromENV()
	}

	globalOrderModInit.Do(func() { globalOrderMod = getOrderModFromENV() })

	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Logf(` Test Execution Seed: %s`, globalOrderMod)
		}
	})
	return globalOrderMod
}

func getOrderModFromENV() testOrderingMod {
	mod, ok := os.LookupEnv(EnvKeyOrderMod)
	if !ok {
		return OrderingAsRandom
	}

	switch testOrderingMod(mod) {
	case OrderingAsDefined:
		return OrderingAsDefined
	case OrderingAsRandom:
		return OrderingAsRandom
	default:
		panic(fmt.Sprintf(`unknown testCase ordering/arrange mod: %s`, mod))
	}
}
