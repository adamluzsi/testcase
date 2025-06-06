package testcase

import (
	"fmt"
	"math/rand"
	"os"
	"sync"

	"go.llib.dev/testcase/internal"
	"go.llib.dev/testcase/internal/environ"
)

func newOrderer(seed int64) orderer {
	switch mod := getGlobalOrderMod(); mod {
	case OrderingAsDefined:
		return nullOrderer{}
	case OrderingAsRandom, undefinedOrdering:
		return randomOrderer{Seed: seed}
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

//---------------------------------------------- Global Test ordering Mod ----------------------------------------------//

var (
	globalOrderMod     testOrderingMod
	globalOrderModInit sync.Once
	_                  = internal.RegisterCacheFlush(func() {
		globalOrderModInit = sync.Once{}
	})
)

func getGlobalOrderMod() testOrderingMod {
	globalOrderModInit.Do(func() { globalOrderMod = getOrderingModFromENV() })
	return globalOrderMod
}

func getOrderingModFromENV() testOrderingMod {
	var (
		mod string
		ok  bool
	)
	for _, envKey := range environ.OrderingKeys() {
		mod, ok = os.LookupEnv(envKey)
		if ok {
			break
		}
	}
	if !ok {
		return OrderingAsRandom
	}
	switch testOrderingMod(mod) {
	case OrderingAsDefined:
		return OrderingAsDefined
	case OrderingAsRandom:
		return OrderingAsRandom
	default:
		panic(fmt.Sprintf(`Unknown testcase ordering/arrange mod: %s\n\nSupported values: %s, %s`, mod, OrderingAsDefined, OrderingAsRandom))
	}
}
