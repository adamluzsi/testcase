package assert_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
)

func SpecWaiter(tb testing.TB) {
	s := testcase.NewSpec(tb)

	var (
		waitTimeout = testcase.Let(s, func(t *testcase.T) time.Duration {
			return time.Millisecond
		})
		helper = testcase.Let(s, func(t *testcase.T) *assert.Waiter {
			return &assert.Waiter{
				Timeout: waitTimeout.Get(t),
			}
		})
	)

	measureDuration := func(fn func()) time.Duration {
		start := time.Now()
		fn()
		stop := time.Now()
		return stop.Sub(start)
	}

	s.Describe(`#Wait`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) {
			helper.Get(t).Wait()
		}

		itShouldNotSpendMuchMoreTimeOnWaitingThanWhatWasDefined := func(s *testcase.Spec) {
			s.Then(`it should around the WaitDuration defined time`, func(t *testcase.T) {
				duration := helper.Get(t).WaitDuration

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
				assert.Must(t).True(min <= avg, `#Wait() should run at least for the duration of WaitDuration`)
				assert.Must(t).True(avg <= max, assert.Message(fmt.Sprintf(`#Wait() shouldn't run more than the WaitDuration + %d%% tolerance`, int(extraTimePercentage*100))))
			})
		}

		s.When(`sleep time is set`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				helper.Get(t).WaitDuration = time.Millisecond
			})

			s.Then(`calling wait will have at least the wait sleep duration`, func(t *testcase.T) {
				assert.Must(t).True(time.Millisecond <= measureDuration(func() { subject(t) }))
			})

			itShouldNotSpendMuchMoreTimeOnWaitingThanWhatWasDefined(s)
		})

		s.When(`sleep time is not set (zero value)`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				var zeroDuration time.Duration
				helper.Get(t).WaitDuration = zeroDuration
			})

			s.Then(`calling wait will have at least the wait sleep duration`, func(t *testcase.T) {
				assert.Must(t).True(measureDuration(func() { subject(t) }) <= time.Millisecond)
			})

			itShouldNotSpendMuchMoreTimeOnWaitingThanWhatWasDefined(s)
		})
	})

	s.Describe(`#While`, func(s *testcase.Spec) {
		cond := testcase.Var[func() bool]{ID: `condition function`}
		counter := testcase.LetValue[int](s, 0)
		duration := testcase.Var[time.Duration]{ID: `condition evaluation duration time`}
		waitTimeout.LetValue(s, time.Millisecond)

		var subject = func(t *testcase.T) {
			helper.Get(t).While(cond.Get(t))
		}

		letCondition := func(s *testcase.Spec, fn func(*testcase.T) bool) {
			counter.LetValue(s, 0)

			cond.Let(s, func(t *testcase.T) func() bool {
				return func() bool {
					counter.Set(t, counter.Get(t)+1)
					time.Sleep(duration.Get(t))
					return fn(t)
				}
			})
		}

		s.When(`the condition never returns with wait no longer needed (true)`, func(s *testcase.Spec) {
			duration.LetValue(s, time.Millisecond)
			letCondition(s, func(t *testcase.T) bool { return true })

			s.And(`wait timeout is shorter that the time it takes to evaluate the condition`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					helper.Get(t).Timeout = time.Duration(t.Random.IntBetween(0, int(duration.Get(t))-1))
				})

				s.Then(`it will execute the condition at least once`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).Equal(1, counter.Get(t))
				})
			})

			s.And(`wait timeout is longer than what it takes to run condition evaluation even multiple times`, func(s *testcase.Spec) {
				duration.LetValue(s, time.Nanosecond)

				s.Before(func(t *testcase.T) {
					helper.Get(t).Timeout = 42 * time.Millisecond
				})

				s.Then(`it will run for as long as the wait timeout duration`, func(t *testcase.T) {
					assert.Must(t).True(helper.Get(t).Timeout <= measureDuration(func() { subject(t) }))
				})

				s.Then(`it will execute the condition multiple times`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).True(1 < counter.Get(t))
				})
			})
		})

		s.When(`the condition quickly returns with done (false)`, func(s *testcase.Spec) {
			duration.LetValue(s, time.Millisecond)
			letCondition(s, func(t *testcase.T) bool { return false })

			s.And(`wait timeout is shorter that the time it takes to evaluate the condition`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					helper.Get(t).Timeout = time.Duration(t.Random.IntBetween(0, int(duration.Get(t))-1))
				})

				s.Then(`it will execute the condition at least once`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).Equal(1, counter.Get(t))
				})
			})

			s.And(`wait timeout is longer than what it takes to run condition evaluation even multiple times`, func(s *testcase.Spec) {
				duration.LetValue(s, time.Nanosecond)
				waitTimeout.LetValue(s, time.Millisecond)

				s.Then(`it will not use up list the time that wait time allows because the condition doesn't need it`, func(t *testcase.T) {
					assert.Must(t).True(measureDuration(func() { subject(t) }) < helper.Get(t).Timeout)
				})

				s.Then(`it will execute the condition only for the required required amount of times`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).Equal(1, counter.Get(t))
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
