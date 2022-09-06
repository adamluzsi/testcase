package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/doubles"
)

func TestRunSuite(t *testing.T) {
	t.Run(`when TB is testing.TB`, func(t *testing.T) {
		sT := &RunContractContract{}
		var tb testing.TB = &doubles.TB{}
		tb = testcase.NewT(tb, testcase.NewSpec(tb))
		testcase.RunSuite(&tb, sT)
		assert.Must(t).True(sT.SpecWasCalled)
		assert.Must(t).True(!sT.TestWasCalled)
		assert.Must(t).True(!sT.BenchmarkWasCalled)
	})

	t.Run(`when TB is *testcase.Spec for *testing.T with #Suite`, func(t *testing.T) {
		s := testcase.NewSpec(t)
		a := &RunContractContract{}
		b := &RunContractContract{}
		testcase.RunSuite(s, a, b)
		s.Finish()
		assert.Must(t).True(a.SpecWasCalled)
		assert.Must(t).True(b.SpecWasCalled)
		assert.Must(t).True(!a.TestWasCalled)
		assert.Must(t).True(!a.BenchmarkWasCalled)
		assert.Must(t).True(!b.TestWasCalled)
		assert.Must(t).True(!b.BenchmarkWasCalled)
	})

	t.Run(`when TB is TBRunner`, func(t *testing.T) {
		var ctb testing.TB = &CustomTB{TB: t}
		contract := &RunContractContract{}
		testcase.RunSuite(&ctb, contract)

		assert.Must(t).True(contract.SpecWasCalled, `because *testing.T is wrapped in the TBRunner`)
		assert.Must(t).True(!contract.TestWasCalled, `because *testing.T is wrapped in the TBRunner`)
		assert.Must(t).True(!contract.BenchmarkWasCalled)
	})
}
func TestRunOpenSuite(t *testing.T) {
	t.Run(`when TB is *testing.T`, func(t *testing.T) {
		sT := &RunContractOpenContract{}
		testcase.RunOpenSuite(t, sT)
		assert.Must(t).True(sT.TestWasCalled)
		assert.Must(t).True(!sT.BenchmarkWasCalled)
	})

	t.Run(`when TB is *testcase.T with *testing.T under the hood`, func(t *testing.T) {
		sT := &RunContractOpenContract{}
		testcase.RunOpenSuite(testcase.NewT(t, nil), sT)
		assert.Must(t).True(sT.TestWasCalled)
		assert.Must(t).True(!sT.BenchmarkWasCalled)
	})

	t.Run(`when TB is *testcase.Spec for *testing.T with #Suite`, func(t *testing.T) {
		s := testcase.NewSpec(t)
		a := &RunContractOpenContract{}
		b := &RunContractOpenContract{}
		testcase.RunOpenSuite(s, a, b)
		s.Finish()
		assert.Must(t).True(a.TestWasCalled)
		assert.Must(t).True(!a.BenchmarkWasCalled)
		assert.Must(t).True(b.TestWasCalled)
		assert.Must(t).True(!b.BenchmarkWasCalled)
	})

	t.Run(`when TB is TBRunner`, func(t *testing.T) {
		var ctb testing.TB = &CustomTB{TB: t}
		contract := &RunContractOpenContract{}
		testcase.RunOpenSuite(&ctb, contract)

		assert.Must(t).True(contract.TestWasCalled, `because *testing.T is wrapped in the TBRunner`)
		assert.Must(t).True(!contract.BenchmarkWasCalled)
	})
}

func BenchmarkTestRunOpenSuite(b *testing.B) {
	b.Run(`when TB is *testing.B`, func(b *testing.B) {
		sB := &RunContractOpenContract{}
		testcase.RunOpenSuite(b, sB)
		assert.Must(b).True(!sB.TestWasCalled)
		assert.Must(b).True(sB.BenchmarkWasCalled)
		b.SkipNow()
	})

	b.Run(`when TB is *testcase.T with *testing.B under the hood`, func(b *testing.B) {
		sT := &RunContractOpenContract{}
		testcase.RunOpenSuite(testcase.NewT(b, nil), sT)
		assert.Must(b).True(!sT.TestWasCalled)
		assert.Must(b).True(sT.BenchmarkWasCalled)
		b.SkipNow()
	})
}

func TestOutput_runContract_fmtStringer(t *testing.T) {
	t.Log("smoke-test")
	testcase.RunSuite(testcase.NewSpec(t), RunContractFmtStringerContract{})
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
