package testcase

import (
	"fmt"
	"testing"
)

// Contract meant to represent a Role Interface Contract.
// A role interface is a static code contract that expresses behavioral expectations as a set of method signatures.
// A role interface used by one or many consumers.
// These consumers often use implicit assumptions about how methods of the role interface behave.
// Using these assumptions makes it possible to simplify the consumer code.
// In testcase convention, instead of relying on implicit assumptions, the developer should create an explicit interface testing suite, in other words, a Contract.
// The code that supplies a role interface then able to import a role interface Contract,
// and confirm if the expected behavior is fulfilled by the implementation.
type Contract interface {
	// Test is the function that assert expected behavioral requirements from a supplier implementation.
	// These behavioral assumptions made by the Consumer in order to simplify and stabilise its own code complexity.
	// Every time a Consumer makes an assumption about the behavior of the role interface supplier,
	// it should be clearly defined it with tests under this functionality.
	Test(*testing.T)
	// Benchmark will help with what to measure.
	// When you define a role interface contract, you should clearly know what performance aspects important for your Consumer.
	// Those aspects should be expressed in a form of Benchmark,
	// so different supplier implementations can be easily A/B tested from this aspect as well.
	Benchmark(*testing.B)
}

// RunContract is a helper function that makes execution one or many Contract easy.
// By using RunContract, you don't have to distinguish between testing or benchmark execution mod.
// It supports *testing.T, *testing.B, *testcase.T, *testcase.Spec and CustomTB test runners.
func RunContract(tb interface{}, contracts ...Contract) {
	for _, c := range contracts {
		c := c
		switch tb := tb.(type) {
		case *testing.T:
			c.Test(tb)

		case *testing.B:
			c.Benchmark(tb)

		case *T:
			RunContract(tb.TB, c)

		case TBRunner:
			c := c
			tb.Run(``, func(tb testing.TB) { RunContract(tb, c) })

		case *Spec:
			tb.Test(``, func(t *T) { RunContract(t, c) })

		default:
			panic(fmt.Errorf(`unknown test runner type: %T`, tb))
		}
	}
}
