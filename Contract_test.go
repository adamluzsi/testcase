package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
)

func TestRunContracts(t *testing.T) {
	t.Run(`when TB is *testing.T`, func(t *testing.T) {
		sT := &RunContractExampleContract{}
		testcase.RunContract(&testing.T{}, sT)
		require.True(t, sT.TestWasCalled)
		require.False(t, sT.BenchmarkWasCalled)
	})

	t.Run(`when TB is *testing.B`, func(t *testing.T) {
		sB := &RunContractExampleContract{}
		testcase.RunContract(&testing.B{}, sB)
		require.False(t, sB.TestWasCalled)
		require.True(t, sB.BenchmarkWasCalled)
	})

	t.Run(`when TB is *testcase.T with *testing.T under the hood`, func(t *testing.T) {
		sT := &RunContractExampleContract{}
		testcase.RunContract(&testcase.T{TB: &testing.T{}}, sT)
		require.True(t, sT.TestWasCalled)
		require.False(t, sT.BenchmarkWasCalled)
	})

	t.Run(`when TB is *testcase.T with *testing.B under the hood`, func(t *testing.T) {
		sT := &RunContractExampleContract{}
		testcase.RunContract(&testcase.T{TB: &testing.B{}}, sT)
		require.False(t, sT.TestWasCalled)
		require.True(t, sT.BenchmarkWasCalled)
	})

	t.Run(`when TB is *testcase.Spec for *testing.T with #Contract`, func(t *testing.T) {
		s := testcase.NewSpec(t)
		a := &RunContractExampleContract{}
		b := &RunContractExampleContract{}
		testcase.RunContract(s, a, b)
		s.Finish()
		require.True(t, a.TestWasCalled)
		require.False(t, a.BenchmarkWasCalled)
		require.True(t, b.TestWasCalled)
		require.False(t, b.BenchmarkWasCalled)
	})

	t.Run(`when TB is TBRunner`, func(t *testing.T) {
		ctb := &CustomTB{TB: t}
		contract := &RunContractExampleContract{}
		testcase.RunContract(ctb, contract)

		require.True(t, contract.TestWasCalled, `because *testing.T is wrapped in the TBRunner`)
		require.False(t, contract.BenchmarkWasCalled)
	})

	t.Run(`when TB is an unknown test runner type`, func(t *testing.T) {
		type UnknownTestRunner struct{}
		require.Panics(t, func() { testcase.RunContract(UnknownTestingTB{}, &RunContractExampleContract{}) })
	})
}

type RunContractExampleContract struct {
	TestWasCalled      bool
	BenchmarkWasCalled bool
}

func (spec *RunContractExampleContract) Test(t *testing.T) {
	spec.TestWasCalled = true
}

func (spec *RunContractExampleContract) Benchmark(b *testing.B) {
	spec.BenchmarkWasCalled = true
}
