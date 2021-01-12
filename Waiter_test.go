package testcase_test

import (
	"fmt"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func SpecWaiter(tb testing.TB) {
	s := testcase.NewSpec(tb)

	var (
		waitTimeout = s.Let(`async tester helper wait timeout`, func(t *testcase.T) interface{} {
			return time.Millisecond
		})
		helper = s.Let(`async tester helper`, func(t *testcase.T) interface{} {
			return &testcase.Waiter{
				WaitTimeout: waitTimeout.Get(t).(time.Duration),
			}
		})
		helperGet = func(t *testcase.T) *testcase.Waiter { return helper.Get(t).(*testcase.Waiter) }
	)

	measureDuration := func(fn func()) time.Duration {
		start := time.Now()
		fn()
		stop := time.Now()
		return stop.Sub(start)
	}

	s.Describe(`#Wait`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) {
			helperGet(t).Wait()
		}

		itShouldNotSpendMuchMoreTimeOnWaitingThanWhatWasDefined := func(s *testcase.Spec) {
			s.Then(`it should around the WaitDuration defined time`, func(t *testcase.T) {
				duration := helperGet(t).WaitDuration

				var (
					samplingCount int
					totalDuration time.Duration
				)

				const extraTimePercentage = 0.30
				extraTime := time.Duration(float64(duration+time.Millisecond) * extraTimePercentage)
				min := duration
				max := duration + extraTime

				for i := 0; i < 42; i++ {
					samplingCount++
					totalDuration += measureDuration(func() { subject(t) })
				}

				avg := totalDuration / time.Duration(samplingCount)
				t.Logf(`min:%s max:%s avg:%s`, min, max, avg)
				require.True(t, min <= avg, `#Wait() should run at least for the duration of WaitDuration`)
				require.True(t, avg <= max, fmt.Sprintf(`#Wait() shouldn't run more than the WaitDuration + %d%% tolerance`, int(extraTimePercentage*100)))
			})
		}

		s.When(`sleep time is set`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				helperGet(t).WaitDuration = time.Millisecond
			})

			s.Then(`calling wait will have at least the wait sleep duration`, func(t *testcase.T) {
				require.True(t, time.Millisecond <= measureDuration(func() { subject(t) }))
			})

			itShouldNotSpendMuchMoreTimeOnWaitingThanWhatWasDefined(s)
		})

		s.When(`sleep time is not set (zero value)`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				var zeroDuration time.Duration
				helperGet(t).WaitDuration = zeroDuration
			})

			s.Then(`calling wait will have at least the wait sleep duration`, func(t *testcase.T) {
				require.True(t, measureDuration(func() { subject(t) }) <= time.Millisecond)
			})

			itShouldNotSpendMuchMoreTimeOnWaitingThanWhatWasDefined(s)
		})
	})

	s.Describe(`#While`, func(s *testcase.Spec) {
		const conditionVN = `condition function`
		var subject = func(t *testcase.T) {
			helperGet(t).While(t.I(conditionVN).(func() bool))
		}

		waitTimeout.LetValue(s, time.Millisecond)

		const conditionCounterVN = conditionVN + ` call counter`
		conditionCounter := func(t *testcase.T) int { return t.I(conditionCounterVN).(int) }

		const conditionEvaluationDurationVN = `condition evaluation duration time`
		s.LetValue(conditionEvaluationDurationVN, 0)
		conditionEvaluationDuration := func(t *testcase.T) time.Duration { return t.I(conditionEvaluationDurationVN).(time.Duration) }

		letCondition := func(s *testcase.Spec, fn func(*testcase.T) bool) {
			s.LetValue(conditionCounterVN, 0)
			s.Let(conditionVN, func(t *testcase.T) interface{} {
				return func() bool {
					t.Let(conditionCounterVN, conditionCounter(t)+1)
					time.Sleep(conditionEvaluationDuration(t))
					return fn(t)
				}
			})
		}

		s.When(`the condition never returns with wait no longer needed (true)`, func(s *testcase.Spec) {
			s.LetValue(conditionEvaluationDurationVN, time.Millisecond)
			letCondition(s, func(t *testcase.T) bool { return true })

			s.And(`wait timeout is shorter that the time it takes to evaluate the condition`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					helperGet(t).WaitTimeout = time.Duration(fixtures.Random.IntBetween(0, int(conditionEvaluationDuration(t))-1))
				})

				s.Then(`it will execute the condition at least once`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, 1, conditionCounter(t))
				})
			})

			s.And(`wait timeout is longer than what it takes to run condition evaluation even multiple times`, func(s *testcase.Spec) {
				s.LetValue(conditionEvaluationDurationVN, time.Nanosecond)

				s.Before(func(t *testcase.T) {
					helperGet(t).WaitTimeout = 42 * time.Millisecond
				})

				s.Then(`it will run for as long as the wait timeout duration`, func(t *testcase.T) {
					require.True(t, helperGet(t).WaitTimeout <= measureDuration(func() { subject(t) }))
				})

				s.Then(`it will execute the condition multiple times`, func(t *testcase.T) {
					subject(t)

					require.True(t, 1 < conditionCounter(t))
				})
			})
		})

		s.When(`the condition quickly returns with done (false)`, func(s *testcase.Spec) {
			s.LetValue(conditionEvaluationDurationVN, time.Millisecond)

			letCondition(s, func(t *testcase.T) bool { return false })

			s.And(`wait timeout is shorter that the time it takes to evaluate the condition`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					helperGet(t).WaitTimeout = time.Duration(fixtures.Random.IntBetween(0, int(conditionEvaluationDuration(t))-1))
				})

				s.Then(`it will execute the condition at least once`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, 1, conditionCounter(t))
				})
			})

			s.And(`wait timeout is longer than what it takes to run condition evaluation even multiple times`, func(s *testcase.Spec) {
				s.LetValue(conditionEvaluationDurationVN, time.Nanosecond)
				waitTimeout.LetValue(s, time.Millisecond)

				s.Then(`it will not use up all the time that wait time allows because the condition doesn't need it`, func(t *testcase.T) {
					require.True(t, measureDuration(func() { subject(t) }) < helperGet(t).WaitTimeout)
				})

				s.Then(`it will execute the condition only for the required required amount of times`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, 1, conditionCounter(t))
				})
			})
		})
	})
}

func TestWaiter(t *testing.T) {
	SpecWaiter(t)
}

func BenchmarkWaiter(b *testing.B) {
	SpecWaiter(b)
}
