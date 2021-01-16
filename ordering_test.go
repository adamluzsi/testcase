package testcase

import (
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNullOrderer_Order(t *testing.T) {
	s := NewSpec(t)
	s.NoSideEffect()

	var (
		orderer = s.Let(`null orderer`, func(t *T) interface{} {
			return nullOrderer{}
		})
	)

	s.Describe(`Order`, func(s *Spec) {
		var (
			originalIDs = s.Let(`original originalIDs`, func(t *T) interface{} {
				var ids []testCase
				for i := 0; i < 42; i++ {
					ids = append(ids, testCase{id: fixtures.Random.String()})
				}
				return ids
			}).EagerLoading(s)
			orderedIDs = s.Let(`ordered originalIDs`, func(t *T) interface{} {
				return copyTestCases(originalIDs.Get(t).([]testCase))
			})
			subject = func(t *T) {
				orderer.Get(t).(nullOrderer).Order(orderedIDs.Get(t).([]testCase))
			}
		)

		s.Test(`.Order should not affect the order of the id list`, func(t *T) {
			subject(t)
			require.Equal(t, originalIDs.Get(t), orderedIDs.Get(t))
		})
	})
}

func TestRandomOrderer_Order(t *testing.T) {
	s := NewSpec(t)
	s.NoSideEffect()

	var (
		seed    = s.Let(`seed`, func(t *T) interface{} { return int64(fixtures.Random.Int()) })
		seedGet = func(t *T) int64 { return seed.Get(t).(int64) }
		orderer = s.Let(`random orderer`, func(t *T) interface{} {
			return randomOrderer{Seed: seedGet(t)}
		})
	)

	s.Describe(`Order`, func(s *Spec) {
		var (
			originalTests = s.Let(`original tests`, func(t *T) interface{} {
				var ids []testCase
				for i := 0; i < 42; i++ {
					ids = append(ids, testCase{id: fixtures.Random.String()})
				}
				return ids
			}).EagerLoading(s)
			originalTestsGet = func(t *T) []testCase { return originalTests.Get(t).([]testCase) }
			orderedTests     = s.Let(`ordered tests`, func(t *T) interface{} {
				return copyTestCases(originalTests.Get(t).([]testCase))
			})
			orderedTestsGet = func(t *T) []testCase { return orderedTests.Get(t).([]testCase) }
			subject         = func(t *T) {
				orderer.Get(t).(randomOrderer).Order(orderedTests.Get(t).([]testCase))
			}
		)

		s.Then(`the order of ids list will be shuffled up`, func(t *T) {
			require.Equal(t, originalTestsGet(t), orderedTestsGet(t), `initially the order is the same`)
			subject(t) // after ordering
			require.NotEqual(t, originalTestsGet(t), orderedTestsGet(t), `after ordering, it should be different`)
		})

		s.Then(`ordering should not affect the length`, func(t *T) {
			subject(t) // after ordering
			require.Equal(t, len(originalTestsGet(t)), len(orderedTestsGet(t)))
		})

		s.Then(`the ordering should not affect the content of the id list`, func(t *T) {
			subject(t)
			require.ElementsMatch(t, originalTestsGet(t), orderedTestsGet(t))
		})

		s.Then(`shuffling should be deterministic and always the same for the same seed`, func(t *T) {
			//l1 := copyTestCases(orderedTestsGet(t))
			subject(t)
			l2 := copyTestCases(orderedTestsGet(t))
			//require.NotEqual(t, l1, l2)

			// reset order for next shuffling
			orderedTests.Set(t, copyTestCases(originalTestsGet(t)))
			subject(t)
			l3 := copyTestCases(orderedTestsGet(t))
			//require.NotEqual(t, l1, l3)
			require.Equal(t, l2, l3, `both outcome of the shuffle should be the same with the same Seed`)
		})

		s.Then(`different seed yield different shuffling`, func(t *T) {
			Retry{Strategy: Waiter{WaitTimeout: time.Second}}.Assert(t, func(tb testing.TB) {
				orderer.Set(t, randomOrderer{Seed: int64(fixtures.Random.Int())})
				subject(t)
				l1 := copyTestCases(orderedTestsGet(t))
				orderer.Set(t, randomOrderer{Seed: int64(fixtures.Random.Int())})
				subject(t)
				l2 := copyTestCases(orderedTestsGet(t))

				require.NotEqual(tb, l1, l2)
				require.ElementsMatch(tb, l1, l2)
			})
		})
	})
}

func copyTestCases(src []testCase) []testCase {
	dst := make([]testCase, len(src))
	copy(dst, src)
	return dst
}
