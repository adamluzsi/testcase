package faultinject

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func TestInitEnabled(t *testing.T) {
	t.Cleanup(func() { enabled.State = false })
	const envKey = "TESTCASE_FAULT_INJECTION"

	s := testcase.NewSpec(t)

	act := func(t *testcase.T) {
		initEnabled()
	}

	s.Before(func(t *testcase.T) { // clean ahead
		testcase.UnsetEnv(t, envKey)
		enabled.State = false
	})

	s.When("no env var is not set", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			testcase.UnsetEnv(t, envKey)
		})

		s.Then("the default strategy is to set enabled to false", func(t *testcase.T) {
			act(t)

			t.Must.False(enabled.State)
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

		s.Then("enabled is set to true", func(t *testcase.T) {
			act(t)

			t.Must.True(enabled.State)
		})
	})

	s.When("env var is set to FALSE/OFF", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			enabled.State = true
			testcase.SetEnv(t, envKey, t.Random.ElementFromSlice([]string{
				"FALSE",
				"false",
				"OFF",
				"off",
			}).(string))
		})

		s.Then("enabled is set to false", func(t *testcase.T) {
			act(t)

			t.Must.False(enabled.State)
		})
	})
}
