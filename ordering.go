package testcase

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/adamluzsi/testcase/internal"
	"github.com/stretchr/testify/require"
)

func newOrderer(tb testing.TB) orderer {
	tb.Helper()
	switch mod := getGlobalOrderMod(tb); mod {
	case OrderingAsDefined:
		return nullOrderer{}
	case OrderingAsRandom, undefinedOrdering:
		return randomOrderer{Seed: getRandomOrderSeed(tb)}
	default:
		panic(fmt.Sprintf(`unknown ordering mod: %s`, mod))
	}
}

type orderer interface {
	Order(tc []func())
}

type testOrderingMod string

const (
	undefinedOrdering testOrderingMod = ``
	OrderingAsDefined testOrderingMod = `defined`
	OrderingAsRandom  testOrderingMod = `random`
)

//------------------------------------------------- order as defined -------------------------------------------------//

type nullOrderer struct{}

func (o nullOrderer) Order([]func()) {}

//-------------------------------------------------- order randomly --------------------------------------------------//

type randomOrderer struct {
	Seed int64
}

func (o randomOrderer) Order(tests []func()) {
	o.rand().Shuffle(len(tests), o.swapFunc(tests))
}

func (o randomOrderer) rand() *rand.Rand {
	return rand.New(rand.NewSource(o.Seed))
}

func (o randomOrderer) swapFunc(tests []func()) func(i int, j int) {
	return func(i, j int) {
		tests[i], tests[j] = tests[j], tests[i]
	}
}

//---------------------------------------------- Test Sorter Random Seed ---------------------------------------------//

func getRandomOrderSeed(tb testing.TB) (_seed int64) {
	tb.Helper()
	tb.Cleanup(func() {
		tb.Helper()
		if tb.Failed() || testing.Verbose() {
			tb.Logf(`Test Random Order Seed: %d`, _seed)
		}
	})

	rawSeed, globalRandomSeedValueSet := os.LookupEnv(EnvKeyOrderSeed)
	if !globalRandomSeedValueSet {
		return time.Now().UnixNano()
	}

	seed, err := strconv.ParseInt(rawSeed, 10, 64)
	require.Nil(tb, err)
	return seed
}

//---------------------------------------------- Global Test Sorter Mod ----------------------------------------------//

var (
	globalOrderMod     testOrderingMod
	globalOrderModInit sync.Once
	_                  = internal.RegisterCacheFlush(func() {
		globalOrderModInit = sync.Once{}
	})
)

func getGlobalOrderMod(tb testing.TB) testOrderingMod {
	tb.Helper()
	globalOrderModInit.Do(func() { globalOrderMod = getOrderModFromENV() })
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
