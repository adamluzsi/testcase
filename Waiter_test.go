package testcase_test

import (
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/adamluzsi/testcase/internal"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestWaiter(t *testing.T) {
	s := testcase.NewSpec(t)

	const waiterVN = `waiter`
	s.Let(waiterVN, func(t *testcase.T) interface{} {
		return &testcase.Waiter{}
	})
	waiter := func(t *testcase.T) *testcase.Waiter {
		return t.I(waiterVN).(*testcase.Waiter)
	}

	measureDuration := func(fn func()) time.Duration {
		start := time.Now()
		fn()
		stop := time.Now()
		return stop.Sub(start)
	}

	s.Describe(`#Wait`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			waiter(t).Wait()
		}

		s.When(`sleep time is set`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				waiter(t).WaitDuration = time.Millisecond
			})

			s.Then(`calling wait will have at least the wait sleep duration`, func(t *testcase.T) {
				require.True(t, time.Millisecond <= measureDuration(func() { subject(t) }))
			})
		})

		s.When(`sleep time is not set (zero value)`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				var zeroDuration time.Duration
				waiter(t).WaitDuration = zeroDuration
			})

			s.Then(`calling wait will have at least the wait sleep duration`, func(t *testcase.T) {
				require.True(t, measureDuration(func() { subject(t) }) <= time.Millisecond)
			})
		})
	})

	s.Describe(`#WaitWhile`, func(s *testcase.Spec) {
		const conditionVN = `condition function`
		var subject = func(t *testcase.T) {
			waiter(t).WaitWhile(t.I(conditionVN).(func() bool))
		}

		s.Before(func(t *testcase.T) {
			waiter(t).WaitTimeout = time.Millisecond
		})

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

		s.When(`when the condition never returns with wait no longer needed (true)`, func(s *testcase.Spec) {
			s.LetValue(conditionEvaluationDurationVN, time.Millisecond)
			letCondition(s, func(t *testcase.T) bool { return true })

			s.And(`wait timeout is shorter that the time it takes to evaluate the condition`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					waiter(t).WaitTimeout = time.Duration(fixtures.Random.IntBetween(0, int(conditionEvaluationDuration(t))-1))
				})

				s.Then(`it will execute the condition at least once`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, 1, conditionCounter(t))
				})
			})

			s.And(`wait timeout is longer than what it takes to run condition evaluation even multiple times`, func(s *testcase.Spec) {
				s.LetValue(conditionEvaluationDurationVN, time.Nanosecond)

				s.Before(func(t *testcase.T) {
					waiter(t).WaitTimeout = time.Millisecond
				})

				s.Then(`it will run for as long as the wait timeout duration`, func(t *testcase.T) {
					require.True(t, waiter(t).WaitTimeout <= measureDuration(func() { subject(t) }))
				})

				s.Then(`it will execute the condition multiple times`, func(t *testcase.T) {
					subject(t)

					require.True(t, 1 < conditionCounter(t))
				})
			})
		})

		s.When(`when the condition quickly returns with done (false)`, func(s *testcase.Spec) {
			s.LetValue(conditionEvaluationDurationVN, time.Millisecond)

			letCondition(s, func(t *testcase.T) bool { return false })

			s.And(`wait timeout is shorter that the time it takes to evaluate the condition`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					waiter(t).WaitTimeout = time.Duration(fixtures.Random.IntBetween(0, int(conditionEvaluationDuration(t))-1))
				})

				s.Then(`it will execute the condition at least once`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, 1, conditionCounter(t))
				})
			})

			s.And(`wait timeout is longer than what it takes to run condition evaluation even multiple times`, func(s *testcase.Spec) {
				s.LetValue(conditionEvaluationDurationVN, time.Nanosecond)

				s.Before(func(t *testcase.T) {
					waiter(t).WaitTimeout = time.Millisecond
				})

				s.Then(`it will not use up all the time that wait time allows because the condition doesn't need it`, func(t *testcase.T) {
					require.True(t, measureDuration(func() { subject(t) }) < waiter(t).WaitTimeout)
				})

				s.Then(`it will execute the condition only for the required required amount of times`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, 1, conditionCounter(t))
				})
			})
		})
	})

	s.Describe(`#Assert`, func(s *testcase.Spec) {
		const assertionsVN = `assert function`

		const tbVN = `TB`
		getTB := func(t *testcase.T) testing.TB { return t.I(tbVN).(testing.TB) }
		s.Let(tbVN, func(t *testcase.T) interface{} {
			return &internal.StubTB{}
		})

		var subject = func(t *testcase.T) {
			waiter(t).Assert(getTB(t), t.I(assertionsVN).(func(testing.TB)))
		}

		s.Before(func(t *testcase.T) {
			waiter(t).WaitTimeout = time.Millisecond
		})

		const assertionCounterVN = assertionsVN + ` call counter`
		conditionCounter := func(t *testcase.T) int { return t.I(assertionCounterVN).(int) }

		const assertionEvaluationDurationVN = `assertion evaluation duration time`
		s.LetValue(assertionEvaluationDurationVN, 0)
		assertionEvaluationDuration := func(t *testcase.T) time.Duration { return t.I(assertionEvaluationDurationVN).(time.Duration) }

		letAssertions := func(s *testcase.Spec, fn func(*testcase.T, testing.TB)) {
			s.LetValue(assertionCounterVN, 0)
			s.Let(assertionsVN, func(t *testcase.T) interface{} {
				return func(tb testing.TB) {
					t.Let(assertionCounterVN, conditionCounter(t)+1)
					time.Sleep(assertionEvaluationDuration(t))
					fn(t, tb)
				}
			})
		}

		s.When(`when the assertion fails`, func(s *testcase.Spec) {
			s.LetValue(assertionEvaluationDurationVN, time.Millisecond)
			letAssertions(s, func(t *testcase.T, tb testing.TB) { tb.Fail() })

			andMultipleAssertionEventSentToTestingTB := func(s *testcase.Spec) {
				s.And(`and multiple assertion event sent to testing.TB`, func(s *testcase.Spec) {
					letAssertions(s, func(t *testcase.T, tb testing.TB) {
						tb.Error(`foo`)
						tb.Errorf(`%s`, `baz`)
						tb.Fatalf(`%s`, `bar`)
						//goland:noinspection GoUnreachableCode
						tb.FailNow() // `never happens`
					})

					s.Let(tbVN, func(t *testcase.T) interface{} {
						ctrl := gomock.NewController(t)
						t.Defer(ctrl.Finish)
						mock := internal.NewMockTB(ctrl)
						mock.EXPECT().Error(gomock.Eq(`foo`))
						mock.EXPECT().Errorf(gomock.Eq(`%s`), gomock.Eq(`baz`))
						mock.EXPECT().Fatalf(gomock.Eq(`%s`), gomock.Eq(`bar`))
						mock.EXPECT().FailNow().Times(0)
						return mock
					})

					s.Then(`all events replied to the passed testing.TB`, func(t *testcase.T) {
						subject(t)
					})
				})
			}

			s.And(`wait timeout is shorter that the time it takes to evaluate the assertions`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					waiter(t).WaitTimeout = time.Duration(fixtures.Random.IntBetween(0, int(assertionEvaluationDuration(t))-1))
				})

				s.Then(`it will execute the assertion at least once`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, 1, conditionCounter(t))
				})

				s.Then(`it will fail the test`, func(t *testcase.T) {
					subject(t)

					require.True(t, getTB(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)
			})

			s.And(`wait timeout is longer than what it takes to run condition evaluation even multiple times`, func(s *testcase.Spec) {
				s.LetValue(assertionEvaluationDurationVN, time.Nanosecond)

				s.Before(func(t *testcase.T) {
					waiter(t).WaitTimeout = time.Millisecond
				})

				s.Then(`it will run for as long as the wait timeout duration`, func(t *testcase.T) {
					require.True(t, waiter(t).WaitTimeout <= measureDuration(func() { subject(t) }))
				})

				s.Then(`it will execute the condition multiple times`, func(t *testcase.T) {
					subject(t)

					require.True(t, 1 < conditionCounter(t))
				})

				s.Then(`it will fail the test`, func(t *testcase.T) {
					subject(t)

					require.True(t, getTB(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)
			})
		})

		s.When(`when the assertion returns with all happy`, func(s *testcase.Spec) {
			s.LetValue(assertionEvaluationDurationVN, time.Millisecond)

			letAssertions(s, func(t *testcase.T, tb testing.TB) {
				// nothing to do, TB then will not fail
			})

			s.And(`wait timeout is shorter that the time it takes to evaluate the condition`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					waiter(t).WaitTimeout = time.Duration(fixtures.Random.IntBetween(0, int(assertionEvaluationDuration(t))-1))
				})

				s.Then(`it will execute the condition at least once`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, 1, conditionCounter(t))
				})

				s.Then(`it will not mark the passed TB as failed`, func(t *testcase.T) {
					subject(t)

					require.False(t, getTB(t).Failed())
				})
			})

			s.And(`wait timeout is longer than what it takes to run condition evaluation even multiple times`, func(s *testcase.Spec) {
				s.LetValue(assertionEvaluationDurationVN, time.Nanosecond)

				s.Before(func(t *testcase.T) {
					waiter(t).WaitTimeout = time.Millisecond
				})

				s.Then(`it will not use up all the time that wait time allows because the condition doesn't need it`, func(t *testcase.T) {
					require.True(t, measureDuration(func() { subject(t) }) < waiter(t).WaitTimeout)
				})

				s.Then(`it will execute the condition only for the required required amount of times`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, 1, conditionCounter(t))
				})

				s.Then(`it will not mark the passed TB as failed`, func(t *testcase.T) {
					subject(t)

					require.False(t, getTB(t).Failed())
				})
			})
		})
	})
}
