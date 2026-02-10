package faultinject

import (
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
)

func TestInitEnabled(t *testing.T) {
	t.Cleanup(func() { state.Enabled = false })
	const envKey = "TESTCASE_FAULTINJECT"

	s := testcase.NewSpec(t)

	act := func(t *testcase.T) {
		initEnabled()
	}

	s.Before(func(t *testcase.T) { // clean ahead
		testcase.UnsetEnv(t, envKey)
		state.Enabled = false
	})

	s.When("no env var is not set", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			testcase.UnsetEnv(t, envKey)
		})

		s.Then("the default strategy is to set state to false", func(t *testcase.T) {
			act(t)

			assert.Must(t).False(state.Enabled)
		})
	})

	s.When("env var is set to TRUE/ON", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			testcase.SetEnv(t, envKey, t.Random.Pick([]string{
				"TRUE",
				"true",
				"1",
			}).(string))
		})

		s.Then("state is set to true", func(t *testcase.T) {
			act(t)

			assert.Must(t).True(state.Enabled)
		})
	})

	s.When("env var is set to FALSE/OFF", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			state.Enabled = true
			testcase.SetEnv(t, envKey, t.Random.Pick([]string{
				"FALSE",
				"false",
				"0",
			}).(string))
		})

		s.Then("state is set to false", func(t *testcase.T) {
			act(t)

			assert.Must(t).False(state.Enabled)
		})
	})
}
