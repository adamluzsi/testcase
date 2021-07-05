package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase/internal"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
)

func TestRunContract(t *testing.T) {
	t.Run(`contract`, func(t *testing.T) {
		t.Run(`when TB is testing.TB`, func(t *testing.T) {
			sT := &RunContractContract{}
			var tb testing.TB = &internal.StubTB{}
			tb = testcase.NewT(tb, testcase.NewSpec(tb))
			testcase.RunContract(tb, sT)
			require.True(t, sT.SpecWasCalled)
			require.False(t, sT.TestWasCalled)
			require.False(t, sT.BenchmarkWasCalled)
		})

		t.Run(`when TB is *testcase.Spec for *testing.T with #Contract`, func(t *testing.T) {
			s := testcase.NewSpec(t)
			a := &RunContractContract{}
			b := &RunContractContract{}
			testcase.RunContract(s, a, b)
			s.Finish()
			require.True(t, a.SpecWasCalled)
			require.True(t, b.SpecWasCalled)
			require.False(t, a.TestWasCalled)
			require.False(t, a.BenchmarkWasCalled)
			require.False(t, b.TestWasCalled)
			require.False(t, b.BenchmarkWasCalled)
		})

		t.Run(`when TB is TBRunner`, func(t *testing.T) {
			ctb := &CustomTB{TB: t}
			contract := &RunContractContract{}
			testcase.RunContract(ctb, contract)

			require.True(t, contract.SpecWasCalled, `because *testing.T is wrapped in the TBRunner`)
			require.False(t, contract.TestWasCalled, `because *testing.T is wrapped in the TBRunner`)
			require.False(t, contract.BenchmarkWasCalled)
		})

		t.Run(`when TB is an unknown test runner type`, func(t *testing.T) {
			type NotTestingTB struct{}
			require.Panics(t, func() { testcase.RunContract(NotTestingTB{}, &RunContractContract{}) })
		})
	})

	t.Run(`open-contract`, func(t *testing.T) {
		t.Run(`when TB is *testing.T`, func(t *testing.T) {
			sT := &RunContractOpenContract{}
			testcase.RunContract(&testing.T{}, sT)
			require.True(t, sT.TestWasCalled)
			require.False(t, sT.BenchmarkWasCalled)
		})

		t.Run(`when TB is *testing.B`, func(t *testing.T) {
			sB := &RunContractOpenContract{}
			testcase.RunContract(&testing.B{}, sB)
			require.False(t, sB.TestWasCalled)
			require.True(t, sB.BenchmarkWasCalled)
		})

		t.Run(`when TB is *testcase.T with *testing.T under the hood`, func(t *testing.T) {
			sT := &RunContractOpenContract{}
			testcase.RunContract(&testcase.T{TB: &testing.T{}}, sT)
			require.True(t, sT.TestWasCalled)
			require.False(t, sT.BenchmarkWasCalled)
		})

		t.Run(`when TB is *testcase.T with *testing.B under the hood`, func(t *testing.T) {
			sT := &RunContractOpenContract{}
			testcase.RunContract(&testcase.T{TB: &testing.B{}}, sT)
			require.False(t, sT.TestWasCalled)
			require.True(t, sT.BenchmarkWasCalled)
		})

		t.Run(`when TB is *testcase.Spec for *testing.T with #Contract`, func(t *testing.T) {
			s := testcase.NewSpec(t)
			a := &RunContractOpenContract{}
			b := &RunContractOpenContract{}
			testcase.RunContract(s, a, b)
			s.Finish()
			require.True(t, a.TestWasCalled)
			require.False(t, a.BenchmarkWasCalled)
			require.True(t, b.TestWasCalled)
			require.False(t, b.BenchmarkWasCalled)
		})

		t.Run(`when TB is TBRunner`, func(t *testing.T) {
			ctb := &CustomTB{TB: t}
			contract := &RunContractOpenContract{}
			testcase.RunContract(ctb, contract)

			require.True(t, contract.TestWasCalled, `because *testing.T is wrapped in the TBRunner`)
			require.False(t, contract.BenchmarkWasCalled)
		})

		t.Run(`when test runner is not valid`, func(t *testing.T) {
			type NotTestingTB struct{}
			require.Panics(t, func() { testcase.RunContract(NotTestingTB{}, &RunContractOpenContract{}) })
		})
	})
}

func TestRunContract_notContractProvided(t *testing.T) {
	type NotContract struct{}
	require.Panics(t, func() { testcase.RunContract(t, NotContract{}) })
}

func TestOutput_runContract_fmtStringer(t *testing.T) {
	t.Log("smoke-test")
	testcase.RunContract(testcase.NewSpec(t), RunContractFmtStringerContract{})
}

type RunContractOpenContract struct {
	TestWasCalled      bool
	BenchmarkWasCalled bool
}

func (c *RunContractOpenContract) Test(t *testing.T) {
	c.TestWasCalled = true
}

func (c *RunContractOpenContract) Benchmark(b *testing.B) {
	c.BenchmarkWasCalled = true
}

type RunContractContract struct {
	SpecWasCalled      bool
	TestWasCalled      bool
	BenchmarkWasCalled bool
}

func (c *RunContractContract) Spec(s *testcase.Spec) {
	c.SpecWasCalled = true
}

func (c *RunContractContract) Test(t *testing.T) {
	c.TestWasCalled = true
}

func (c *RunContractContract) Benchmark(b *testing.B) {
	c.BenchmarkWasCalled = true
}

type RunContractFmtStringerContract struct{}

func (c RunContractFmtStringerContract) String() string { return "Hello, world!" }
func (c RunContractFmtStringerContract) Spec(s *testcase.Spec) {
	s.Test(``, func(t *testcase.T) { t.Log("!dlrow ,olleH") })
}
