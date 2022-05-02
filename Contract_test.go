package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
)

func TestRunContract(t *testing.T) {
	t.Run(`when TB is testing.TB`, func(t *testing.T) {
		sT := &RunContractContract{}
		var tb testing.TB = &testcase.StubTB{}
		tb = testcase.NewT(tb, testcase.NewSpec(tb))
		testcase.RunContract(tb, sT)
		assert.Must(t).True(sT.SpecWasCalled)
		assert.Must(t).True(!sT.TestWasCalled)
		assert.Must(t).True(!sT.BenchmarkWasCalled)
	})

	t.Run(`when TB is *testcase.Spec for *testing.T with #Contract`, func(t *testing.T) {
		s := testcase.NewSpec(t)
		a := &RunContractContract{}
		b := &RunContractContract{}
		testcase.RunContract(s, a, b)
		s.Finish()
		assert.Must(t).True(a.SpecWasCalled)
		assert.Must(t).True(b.SpecWasCalled)
		assert.Must(t).True(!a.TestWasCalled)
		assert.Must(t).True(!a.BenchmarkWasCalled)
		assert.Must(t).True(!b.TestWasCalled)
		assert.Must(t).True(!b.BenchmarkWasCalled)
	})

	t.Run(`when TB is TBRunner`, func(t *testing.T) {
		ctb := &CustomTB{TB: t}
		contract := &RunContractContract{}
		testcase.RunContract(ctb, contract)

		assert.Must(t).True(contract.SpecWasCalled, `because *testing.T is wrapped in the TBRunner`)
		assert.Must(t).True(!contract.TestWasCalled, `because *testing.T is wrapped in the TBRunner`)
		assert.Must(t).True(!contract.BenchmarkWasCalled)
	})

	t.Run(`when TB is an unknown test runner type`, func(t *testing.T) {
		type NotTestingTB struct{}
		assert.Must(t).Panic(func() { testcase.RunContract(NotTestingTB{}, &RunContractContract{}) })
	})
}
func TestRunOpenContract(t *testing.T) {
	t.Run(`when TB is *testing.T`, func(t *testing.T) {
		sT := &RunContractOpenContract{}
		testcase.RunOpenContract(&testing.T{}, sT)
		assert.Must(t).True(sT.TestWasCalled)
		assert.Must(t).True(!sT.BenchmarkWasCalled)
	})

	t.Run(`when TB is *testing.B`, func(t *testing.T) {
		sB := &RunContractOpenContract{}
		testcase.RunOpenContract(&testing.B{}, sB)
		assert.Must(t).True(!sB.TestWasCalled)
		assert.Must(t).True(sB.BenchmarkWasCalled)
	})

	t.Run(`when TB is *testcase.T with *testing.T under the hood`, func(t *testing.T) {
		sT := &RunContractOpenContract{}
		testcase.RunOpenContract(&testcase.T{TB: &testing.T{}}, sT)
		assert.Must(t).True(sT.TestWasCalled)
		assert.Must(t).True(!sT.BenchmarkWasCalled)
	})

	t.Run(`when TB is *testcase.T with *testing.B under the hood`, func(t *testing.T) {
		sT := &RunContractOpenContract{}
		testcase.RunOpenContract(&testcase.T{TB: &testing.B{}}, sT)
		assert.Must(t).True(!sT.TestWasCalled)
		assert.Must(t).True(sT.BenchmarkWasCalled)
	})

	t.Run(`when TB is *testcase.Spec for *testing.T with #Contract`, func(t *testing.T) {
		s := testcase.NewSpec(t)
		a := &RunContractOpenContract{}
		b := &RunContractOpenContract{}
		testcase.RunOpenContract(s, a, b)
		s.Finish()
		assert.Must(t).True(a.TestWasCalled)
		assert.Must(t).True(!a.BenchmarkWasCalled)
		assert.Must(t).True(b.TestWasCalled)
		assert.Must(t).True(!b.BenchmarkWasCalled)
	})

	t.Run(`when TB is TBRunner`, func(t *testing.T) {
		ctb := &CustomTB{TB: t}
		contract := &RunContractOpenContract{}
		testcase.RunOpenContract(ctb, contract)

		assert.Must(t).True(contract.TestWasCalled, `because *testing.T is wrapped in the TBRunner`)
		assert.Must(t).True(!contract.BenchmarkWasCalled)
	})

	t.Run(`when test runner is not valid`, func(t *testing.T) {
		type NotTestingTB struct{}
		assert.Must(t).Panic(func() { testcase.RunOpenContract(NotTestingTB{}, &RunContractOpenContract{}) })
	})
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
