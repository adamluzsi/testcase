package testcase

import (
	"fmt"
	"math/rand"
	"testing"
)

func newOrderer(tb testing.TB, mod testOrderingMod) orderer {
	switch mod {
	case OrderingAsDefined:
		panic(`NotImplemented`)

	case OrderingAsRandom, undefinedOrdering:
		return &randomOrderer{Seed: getGlobalRandomOrderSeed(tb)}

	default:
		panic(fmt.Sprintf(`unknown ordering mod: %s`, mod))
	}
}

type orderer interface {
	Order(tb testing.TB, ids []string)
}

type testOrderingMod string

const (
	undefinedOrdering testOrderingMod = ``
	OrderingAsDefined testOrderingMod = `defined`
	OrderingAsRandom  testOrderingMod = `random`
)

//------------------------------------------------- order as defined -------------------------------------------------//

type nullOrderer struct{}

func (o nullOrderer) Order(testing.TB, []string) {}

//-------------------------------------------------- order randomly --------------------------------------------------//

type randomOrderer struct {
	Seed int64
}

func (o randomOrderer) Order(tb testing.TB, ids []string) {
	o.rand().Shuffle(len(ids), o.swapFunc(ids))
}

func (o randomOrderer) rand() *rand.Rand {
	source := rand.NewSource(o.Seed)
	return rand.New(source)
}

func (o randomOrderer) swapFunc(ids []string) func(i int, j int) {
	return func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	}
}
