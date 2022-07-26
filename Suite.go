package testcase

import (
	"fmt"
	"testing"

	"github.com/adamluzsi/testcase/internal"
)

// Suite meant to represent a testing suite.
// A test Suite is a collection of test cases.
// In a test suite, the test cases are organized in a logical order.
// A Suite is a great tool to define interface testing suites (contracts).
type Suite interface {
	// Spec defines the tests on the received *Spec object.
	Spec(s *Spec)
}

// OpenSuite is a testcase independent testing suite interface standard.
type OpenSuite interface {
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

// RunSuite is a helper function that makes execution one or many Suite easy.
// By using RunSuite, you don't have to distinguish between testing or benchmark execution mod.
// It supports *testing.T, *testing.B, *testcase.T, *testcase.Spec and CustomTB test runners.
func RunSuite(tb any, contracts ...Suite) {
	if tb, ok := tb.(helper); ok {
		tb.Helper()
	}
	s := getSuiteSpec(tb)
	defer s.Finish()
	for _, c := range contracts {
		c := c
		name := getSuiteName(c)
		s.Context(name, c.Spec, Group(name))
	}
}

func RunOpenSuite(tb any, contracts ...OpenSuite) {
	if tb, ok := tb.(helper); ok {
		tb.Helper()
	}
	s := getSuiteSpec(tb)
	defer s.Finish()
	for _, c := range contracts {
		RunSuite(s, OpenSuiteAdapter{OpenSuite: c})
	}
}

type OpenSuiteAdapter struct{ OpenSuite }

func (c OpenSuiteAdapter) Spec(s *Spec) { c.runOpenSuite(s.testingTB, c.OpenSuite) }

func (c OpenSuiteAdapter) runOpenSuite(tb testing.TB, contract OpenSuite) {
	switch tb := tb.(type) {
	case *T:
		c.runOpenSuite(tb.TB, c)
	case *testing.T:
		c.Test(tb)
	case *testing.B:
		c.Benchmark(tb)
	case TBRunner:
		tb.Run(getSuiteName(c), func(tb testing.TB) { RunOpenSuite(tb, c) })
	default:
		panic(fmt.Errorf(`unknown testing.TB: %T`, tb))
	}
}

func getSuiteSpec(tb interface{}) *Spec {
	switch v := tb.(type) {
	case *Spec:
		return v
	case testing.TB:
		return NewSpec(v)
	default:
		panic(fmt.Errorf(`unknown testing.TB: %T`, v))
	}
}

func getSuiteName(c interface{}) (name string) {
	defer func() { name = escapeName(name) }()
	switch c := c.(type) {
	case interface{ Name() string }:
		return c.Name()
	default:
		return internal.SymbolicName(c)
	}
}
