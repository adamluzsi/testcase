package testcase_test

import (
	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
	"strconv"
	"testing"
)

func TestSpec_seed_T_Random(t *testing.T) {
	t.Run("random values are unique across the testing scenarios", func(t *testing.T) {
		testcase.UnsetEnv(t, "TESTCASE_SEED")

		assert.Equal(t, 4, len(runSeedScenarios(t)))
	})
	t.Run("random values are unique for each spec instance", func(t *testing.T) {
		testcase.UnsetEnv(t, "TESTCASE_SEED")

		assert.NotEqual(t, runSeedScenarios(t), runSeedScenarios(t))
	})
	t.Run("using the TESTCASE_SEED env variable allows us to get back the same random values", func(t *testing.T) {
		testcase.SetEnv(t, "TESTCASE_SEED", "8426361600145010042")

		tbWithDeterministicTestNames := func() *doubles.TB {
			var offset int
			return &doubles.TB{
				StubNameFunc: func() string {
					offset++
					return strconv.Itoa(offset)
				},
			}
		}

		assert.Equal(t,
			runSeedScenarios(tbWithDeterministicTestNames()),
			runSeedScenarios(tbWithDeterministicTestNames()))
	})
}

func runSeedScenarios(tb testing.TB) map[string]struct{} {
	s := testcase.NewSpec(tb)
	s.Sequential()

	var values = make(map[string]struct{})

	s.Test("", func(t *testcase.T) {
		values[t.Random.UUID()] = struct{}{}
	})

	s.Test("", func(t *testcase.T) {
		values[t.Random.UUID()] = struct{}{}
	})

	s.Context("", func(s *testcase.Spec) {
		s.Test("", func(t *testcase.T) {
			values[t.Random.UUID()] = struct{}{}
		})

		s.Test("", func(t *testcase.T) {
			values[t.Random.UUID()] = struct{}{}
		})
	})

	s.Finish()
	return values
}
