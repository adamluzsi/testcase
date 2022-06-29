package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal"
)

func TestLogger(t *testing.T) {
	t.Run("when logger is not provided", func(t *testing.T) {
		tb := &testcase.StubTB{}
		s := testcase.NewSpec(tb)
		s.Test("Log", func(t *testcase.T) { t.Log("Log") })
		s.Test("Logf", func(t *testcase.T) { t.Logf("Logf %s", "v") })
		s.Test("Error", func(t *testcase.T) { t.Error("Error") })
		s.Test("Errorf", func(t *testcase.T) { t.Errorf("Errorf %s", "v") })
		s.Test("Fatal", func(t *testcase.T) { internal.Recover(func() { t.Fatal("Fatal") }) })
		s.Test("Fatalf", func(t *testcase.T) { internal.Recover(func() { t.Fatalf("Fatalf %s", "v") }) })
		s.Test("Skip", func(t *testcase.T) { internal.Recover(func() { t.Skip("Skip") }) })
		s.Test("Skipf", func(t *testcase.T) { internal.Recover(func() { t.Skipf("Skipf %s", "v") }) })
		s.Finish()

		it := assert.MakeIt(t)
		it.Should.Contain(tb.Logs.String(), "Log")
		it.Should.Contain(tb.Logs.String(), "Logf v")
		it.Should.Contain(tb.Logs.String(), "Error")
		it.Should.Contain(tb.Logs.String(), "Errorf v")
		it.Should.Contain(tb.Logs.String(), "Fatal")
		it.Should.Contain(tb.Logs.String(), "Fatalf v")
		it.Should.Contain(tb.Logs.String(), "Skip")
		it.Should.Contain(tb.Logs.String(), "Skipf v")
	})

	t.Run("when logger is provided", func(t *testing.T) {
		tb := &testcase.StubTB{}
		logger := &testcase.StubTB{}
		lv := testcase.Var[testcase.Logger]{
			ID:   "asdf",
			Init: func(t *testcase.T) testcase.Logger { return logger },
		}
		s := testcase.NewSpec(tb, testcase.WithLogger(lv))
		s.Test("Log", func(t *testcase.T) { t.Log("Log") })
		s.Test("Logf", func(t *testcase.T) { t.Logf("Logf %s", "v") })
		s.Test("Error", func(t *testcase.T) { t.Error("Error") })
		s.Test("Errorf", func(t *testcase.T) { t.Errorf("Errorf %s", "v") })
		s.Test("Fatal", func(t *testcase.T) { internal.Recover(func() { t.Fatal("Fatal") }) })
		s.Test("Fatalf", func(t *testcase.T) { internal.Recover(func() { t.Fatalf("Fatalf %s", "v") }) })
		s.Test("Skip", func(t *testcase.T) { internal.Recover(func() { t.Skip("Skip") }) })
		s.Test("Skipf", func(t *testcase.T) { internal.Recover(func() { t.Skipf("Skipf %s", "v") }) })
		s.Finish()

		it := assert.MakeIt(t)
		// testing.TB
		it.Should.NotContain(tb.Logs.String(), "Log")
		it.Should.NotContain(tb.Logs.String(), "Logf v")
		it.Should.NotContain(tb.Logs.String(), "Error")
		it.Should.NotContain(tb.Logs.String(), "Errorf v")
		it.Should.NotContain(tb.Logs.String(), "Fatal")
		it.Should.NotContain(tb.Logs.String(), "Fatalf v")
		it.Should.NotContain(tb.Logs.String(), "Skip")
		it.Should.NotContain(tb.Logs.String(), "Skipf v")
		// testcase.Logger
		it.Should.Contain(logger.Logs.String(), "Log")
		it.Should.Contain(logger.Logs.String(), "Logf v")
		it.Should.Contain(logger.Logs.String(), "Error")
		it.Should.Contain(logger.Logs.String(), "Errorf v")
		it.Should.Contain(logger.Logs.String(), "Fatal")
		it.Should.Contain(logger.Logs.String(), "Fatalf v")
		it.Should.Contain(logger.Logs.String(), "Skip")
		it.Should.Contain(logger.Logs.String(), "Skipf v")
	})
}
