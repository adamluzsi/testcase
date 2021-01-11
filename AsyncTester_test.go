package testcase_test

import (
	"github.com/adamluzsi/testcase/internal/mocks"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/adamluzsi/testcase/internal"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAsyncTester(t *testing.T) {
	SpecAsyncTester(t)
}

func BenchmarkAsyncTester(b *testing.B) {
	SpecAsyncTester(b)
}

func SpecAsyncTester(tb testing.TB) {
	s := testcase.NewSpec(tb)

	var (
		waitTimeout = s.Let(`async tester helper wait timeout`, func(t *testcase.T) interface{} {
			return time.Millisecond
		})
		helper = s.Let(`async tester helper`, func(t *testcase.T) interface{} {
			return &testcase.AsyncTester{
				Waiter: testcase.Waiter{WaitTimeout: waitTimeout.Get(t).(time.Duration)},
			}
		})
		helperGet = func(t *testcase.T) *testcase.AsyncTester { return helper.Get(t).(*testcase.AsyncTester) }
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

				const extraTimePercentage = 0.2
				extraTime := time.Duration(float64(duration+time.Millisecond) * extraTimePercentage)

				min := duration
				max := duration + extraTime
				actual := measureDuration(func() { subject(t) })
				//t.Logf(`duration: %d [min:%d max:%d]`, actual, min, max)

				require.True(t, min <= actual, `#Wait() should run at least for the duration of WaitDuration`)
				require.True(t, actual <= max, `#Wait() shouldn't run more than the WaitDuration + 20% tolerance`)
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

	s.Describe(`#WaitWhile`, func(s *testcase.Spec) {
		const conditionVN = `condition function`
		var subject = func(t *testcase.T) {
			helperGet(t).WaitWhile(t.I(conditionVN).(func() bool))
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
					helperGet(t).WaitTimeout = time.Millisecond
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

	s.Describe(`#Assert`, func(s *testcase.Spec) {
		var (
			TB      = s.Let(`TB`, func(t *testcase.T) interface{} { return &internal.StubTB{} })
			tbGet   = func(t *testcase.T) testing.TB { return TB.Get(t).(testing.TB) }
			blk     = s.Let(`assert function`, func(t *testcase.T) interface{} { return func(testing.TB) {} })
			blkGet  = func(t *testcase.T) func(testing.TB) { return blk.Get(t).(func(testing.TB)) }
			subject = func(t *testcase.T) {
				helperGet(t).Assert(tbGet(t), blkGet(t))
			}
		)

		var (
			blkCounter     = s.LetValue(blk.Name+` call blkCounter`, 0)
			blkCounterGet  = func(t *testcase.T) int { return blkCounter.Get(t).(int) }
			blkDuration    = s.LetValue(`assertion evaluation duration time`, time.Duration(0))
			blkDurationGet = func(t *testcase.T) time.Duration { return blkDuration.Get(t).(time.Duration) }
			blkLet         = func(s *testcase.Spec, fn func(*testcase.T, testing.TB)) {
				blkCounterInc := func(t *testcase.T) { blkCounter.Set(t, blkCounter.Get(t).(int)+1) }

				blk.Let(s, func(t *testcase.T) interface{} {
					return func(tb testing.TB) {
						blkCounterInc(t)
						time.Sleep(blkDurationGet(t))
						fn(t, tb)
					}
				})
			}
		)

		s.When(`the assertion fails`, func(s *testcase.Spec) {
			blkLet(s, func(t *testcase.T, tb testing.TB) { tb.Fail() })

			andMultipleAssertionEventSentToTestingTB := func(s *testcase.Spec) {
				s.And(`and multiple assertion event sent to testing.TB`, func(s *testcase.Spec) {
					cuCounter := s.LetValue(`cleanup blkCounter`, 0)

					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Cleanup(func() { cuCounter.Set(t, cuCounter.Get(t).(int)+1) })
						tb.Error(`foo`)
						tb.Errorf(`%s`, `baz`)
						tb.Fatalf(`%s`, `bar`)
						//goland:noinspection GoUnreachableCode
						tb.FailNow() // `never happens`
					})

					TB.Let(s, func(t *testcase.T) interface{} {
						ctrl := gomock.NewController(t)
						t.Defer(ctrl.Finish)
						mock := mocks.NewMockTB(ctrl)
						mock.EXPECT().Cleanup(gomock.Any()).Do(func(f func()) { f() }).AnyTimes()
						mock.EXPECT().Error(gomock.Eq(`foo`))
						mock.EXPECT().Errorf(gomock.Eq(`%s`), gomock.Eq(`baz`))
						mock.EXPECT().Fatalf(gomock.Eq(`%s`), gomock.Eq(`bar`))
						mock.EXPECT().FailNow().Times(0)
						return mock
					})

					s.Then(`all events replied to the passed testing.TB`, func(t *testcase.T) {
						subject(t)
					})

					s.Then(`cleanup is forwarded regardless the failed error`, func(t *testcase.T) {
						subject(t)

						require.Greater(t, cuCounter.Get(t), 0)
					})
				})
			}

			s.And(`wait timeout is shorter that the time it takes to evaluate the assertions`, func(s *testcase.Spec) {
				blkDuration.LetValue(s, time.Millisecond)
				waitTimeout.LetValue(s, time.Nanosecond)

				s.Then(`it will execute the assertion at least once`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, 1, blkCounterGet(t))
				})

				s.Then(`it will fail the test`, func(t *testcase.T) {
					subject(t)

					require.True(t, tbGet(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)
			})

			s.And(`wait timeout is longer than what it takes to run condition evaluation even multiple times`, func(s *testcase.Spec) {
				blkDuration.LetValue(s, time.Nanosecond)
				waitTimeout.LetValue(s, time.Millisecond)

				s.Then(`it will run for as long as the wait timeout duration`, func(t *testcase.T) {
					require.True(t, helperGet(t).WaitTimeout <= measureDuration(func() { subject(t) }))
				})

				s.Then(`it will execute the condition multiple times`, func(t *testcase.T) {
					subject(t)

					require.True(t, 1 < blkCounterGet(t))
				})

				s.Then(`it will fail the test`, func(t *testcase.T) {
					subject(t)

					require.True(t, tbGet(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)

				s.And(`it fails with FailNow`, func(s *testcase.Spec) {
					hasRun := s.LetValue(`hasRun`, false)

					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Cleanup(func() { hasRun.Set(t, true) })
						tb.FailNow()
					})

					s.Then(`it will fail the test`, func(t *testcase.T) {
						internal.InGoroutine(func() { subject(t) })

						require.True(t, tbGet(t).Failed())
					})

					s.Then(`it will ensure that Cleanup was executed`, func(t *testcase.T) {
						internal.InGoroutine(func() { subject(t) })

						require.True(t, hasRun.Get(t).(bool))
					})
				})
			})
		})

		s.When(`the assertion returns with all happy`, func(s *testcase.Spec) {
			blkLet(s, func(t *testcase.T, tb testing.TB) {
				// nothing to do, TB then will not fail // tb.Success
			})

			andMultipleAssertionEventSentToTestingTB := func(s *testcase.Spec) {
				s.And(`and multiple assertion event sent to testing.TB`, func(s *testcase.Spec) {
					cuCounter := s.LetValue(`cleanup blkCounter`, 0)

					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Log(`foo`)
						tb.Logf(`%s - %s`, `bar`, `baz`)
						tb.Cleanup(func() { cuCounter.Set(t, cuCounter.Get(t).(int)+1) })
					})

					TB.Let(s, func(t *testcase.T) interface{} {
						ctrl := gomock.NewController(t)
						t.Defer(ctrl.Finish)
						mock := mocks.NewMockTB(ctrl)
						mock.EXPECT().Log(gomock.Eq(`foo`))
						mock.EXPECT().Logf(gomock.Eq(`%s - %s`), gomock.Eq(`bar`), gomock.Eq(`baz`))
						mock.EXPECT().Cleanup(gomock.Any()).Do(func(f func()) { f() }).AnyTimes()
						return mock
					})

					s.Then(`all events replied to the passed testing.TB`, func(t *testcase.T) {
						subject(t)
					})

					s.Then(`cleanup is forwarded`, func(t *testcase.T) {
						subject(t)

						require.Greater(t, cuCounter.Get(t), 0)
					})
				})
			}

			s.And(`wait timeout is shorter that the time it takes to evaluate the condition`, func(s *testcase.Spec) {
				waitTimeout.LetValue(s, time.Nanosecond)
				blkDuration.LetValue(s, time.Millisecond)

				s.Then(`it will execute the condition at least once`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, 1, blkCounterGet(t))
				})

				s.Then(`it will not mark the passed TB as failed`, func(t *testcase.T) {
					subject(t)

					require.False(t, tbGet(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)
			})

			s.And(`wait timeout is longer than what it takes to run condition evaluation even multiple times`, func(s *testcase.Spec) {
				waitTimeout.LetValue(s, time.Second)
				blkDuration.LetValue(s, time.Duration(0))

				s.Then(`it will not use up all the time that wait time allows because the condition doesn't need it`, func(t *testcase.T) {
					require.True(t, measureDuration(func() { subject(t) }) < helperGet(t).WaitTimeout)
				})

				s.Then(`it will execute the condition only for the required required amount of times`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, 1, blkCounterGet(t))
				})

				s.Then(`it will not mark the passed TB as failed`, func(t *testcase.T) {
					subject(t)

					require.False(t, tbGet(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)

				s.Context(`but it will fail during the Cleanup`, func(s *testcase.Spec) {
					TB.Let(s, func(t *testcase.T) interface{} {
						return mocks.NewWithDefaults(t, func(mock *mocks.MockTB) {
							mock.EXPECT().Cleanup(gomock.Any()).Do(func(f func()) { f() })
							mock.EXPECT().FailNow()
						})
					})

					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Cleanup(func() {
							tb.Logf(`I'm running and I'm pointing to %T`, tb)
							tb.FailNow()
						})
					})

					s.Then(`then cleanup is replied to the test subject`, func(t *testcase.T) {
						subject(t) // assertion in the TB mock
					})
				})

				s.And(`assertion fails a few times but then yields success`, func(s *testcase.Spec) {
					waitTimeout.LetValue(s, time.Hour)

					TB.Let(s, func(t *testcase.T) interface{} {
						m := mocks.New(t)
						t.Log(TB.Name + ` will receive the last stack of Cleanup`)
						t.Log(`it is a design decision at the moment that the last stack of deferred Cleanup is passed to the parent`)
						m.EXPECT().Cleanup(gomock.Any()).Times(3)
						t.Log(TB.Name + ` will not receive FailNow since the Assert Block succeeds before the wait timeout would have been reached`)
						m.EXPECT().FailNow().Times(0) // Flaky Warning! // TestAsyncTester/#Assert/#21
						return m
					})

					var (
						cleanups       = s.Let(`cleanups`, func(t *testcase.T) interface{} { return []string{} })
						cleanupsGet    = func(t *testcase.T) []string { return cleanups.Get(t).([]string) }
						cleanupsAppend = func(t *testcase.T, v ...string) {
							cleanups.Set(t, append(cleanups.Get(t).([]string), v...))
						}
					)
					blkLet(s, func(t *testcase.T, tb testing.TB) {
						t.Log(`ent`)
						tb.Cleanup(func() { cleanupsAppend(t, `foo`) })
						tb.Cleanup(func() { cleanupsAppend(t, `bar`) })
						tb.Cleanup(func() { cleanupsAppend(t, `baz`) })

						t.Log(`in`)
						// fail happens after the cleanups intentionally
						t.Log(`blkCounterGet`, blkCounterGet(t))
						if i := blkCounterGet(t); i < 3 {
							t.Log(`err`)
							tb.FailNow()
						}
						t.Log(`out`)
					})

					s.Then(`failed runs cleanup after themselves`, func(t *testcase.T) {
						subject(t) // expectations in in the TB input as mock

						expected := []string{
							`baz`, `bar`, `foo`, // block runs first
							`baz`, `bar`, `foo`, // block runs for the second time
						}

						require.Equal(t, expected, cleanupsGet(t))
					})
				})
			})
		})
	})
}

func TestAsyncTester_Assert_failsOnceButThenPass(t *testing.T) {
	w := testcase.AsyncTester{
		Waiter: testcase.Waiter{
			WaitDuration: 0,
			WaitTimeout:  42 * time.Second,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockTB(ctrl)
	m.EXPECT().Cleanup(gomock.Any()).Do(func(f func()) { f() }).AnyTimes()

	var counter int
	var times int
	w.Assert(m, func(tb testing.TB) {
		// run 42 times
		// 41 times it will fail but create cleanup
		// on the last it should pass
		//
		tb.Cleanup(func() { counter++ })
		if 41 <= times {
			return
		}
		times++
		tb.Fail()
	})

	require.Equal(t, 42, counter)
}
