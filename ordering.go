package testcase

import (
	"fmt"
	"math/rand"
	"testing"
)

func newOrderer(tb testing.TB, mod testOrderingMod) orderer {
	switch mod {
	case OrderingAsDefined:
		return nullOrderer{}

	case OrderingAsRandom, undefinedOrdering:
		return randomOrderer{Seed: getGlobalRandomOrderSeed(tb)}

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
