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
	Order(tc []testCase)
}

type testOrderingMod string

const (
	undefinedOrdering testOrderingMod = ``
	OrderingAsDefined testOrderingMod = `defined`
	OrderingAsRandom  testOrderingMod = `random`
)

//------------------------------------------------- order as defined -------------------------------------------------//

type nullOrderer struct{}

func (o nullOrderer) Order([]testCase) {}

//-------------------------------------------------- order randomly --------------------------------------------------//

type randomOrderer struct {
	Seed int64
}

func (o randomOrderer) Order(tcs []testCase) {
	var (
		tests = make([]testCase, 0, len(tcs))
		index = make(map[string][]testCase)
		ids   = make([]string, 0)
	)

	for _, tc := range tcs {
		if _, ok := index[tc.id]; !ok {
			ids = append(ids, tc.id)
		}

		index[tc.id] = append(index[tc.id], tc)
	}

	o.rand().Shuffle(len(ids), o.swapFunc(ids))

	for _, id := range ids {
		if tcs, ok := index[id]; ok {
			tests = append(tests, tcs...)
		}
	}

	for i, tc := range tests {
		tcs[i] = tc
	}
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
