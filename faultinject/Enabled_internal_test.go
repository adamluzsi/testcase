package faultinject

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
)

func TestInitEnabled(t *testing.T) {
	const envKey = "TESTCASE_FAULT_INJECTION"
	assert.True(t, Enabled, "expected default state of Enabled should be true")

	s := testcase.NewSpec(t)

	act := func(t *testcase.T) {
		initEnabled()
	}

	s.Before(func(t *testcase.T) { // clean ahead
		testcase.UnsetEnv(t, envKey)
		ForTest(t, false)
	})

	s.When("no env var is not set", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			testcase.UnsetEnv(t, envKey)
		})

		s.Then("the default strategy is to set Enabled to true", func(t *testcase.T) {
			act(t)

			t.Must.True(Enabled)
		})
	})

	s.When("env var is set to TRUE/ON", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			testcase.SetEnv(t, envKey, t.Random.ElementFromSlice([]string{
				"TRUE",
				"true",
				"on",
				"ON",
			}).(string))
		})

		s.Then("Enabled is set to true", func(t *testcase.T) {
			act(t)

			t.Must.True(Enabled)
		})
	})

	s.When("env var is set to FALSE/OFF", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			Enabled = true
			testcase.SetEnv(t, envKey, t.Random.ElementFromSlice([]string{
				"FALSE",
				"false",
				"OFF",
				"off",
			}).(string))
		})

		s.Then("Enabled is set to false", func(t *testcase.T) {
			act(t)

			t.Must.False(Enabled)
		})
	})
}
