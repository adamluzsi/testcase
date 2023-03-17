package assert_test

import (
	"fmt"
	"github.com/adamluzsi/testcase/let"
	"testing"
	"time"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/random"
	"github.com/adamluzsi/testcase/sandbox"

	"github.com/adamluzsi/testcase"
)

func TestEventually(t *testing.T) {
	SpecEventually(t)
}

func BenchmarkEventually(b *testing.B) {
	SpecEventually(b)
}

func SpecEventually(tb testing.TB) {
	s := testcase.NewSpec(tb)

	var (
		strategyWillRetry = testcase.Var[bool]{ID: `retry strategy will retry`}
		strategyStub      = testcase.Let(s, func(t *testcase.T) *stubRetryStrategy {
			return &stubRetryStrategy{ShouldRetry: strategyWillRetry.Get(t)}
		})
		helper = testcase.Let(s, func(t *testcase.T) *assert.Eventually {
			return &assert.Eventually{
				RetryStrategy: strategyStub.Get(t),
			}
		})
	)

	s.Describe(`.Assert`, func(s *testcase.Spec) {
		var (
			stubTB = testcase.Let(s, func(t *testcase.T) *doubles.TB { return &doubles.TB{} })
			blk    = testcase.Let(s, func(t *testcase.T) func(assert.It) { return func(it assert.It) {} })
		)
		act := func(t *testcase.T) {
			helper.Get(t).Assert(stubTB.Get(t), blk.Get(t))
		}

		var (
			blkCounter     = testcase.LetValue(s, 0)
			blkCounterGet  = func(t *testcase.T) int { return blkCounter.Get(t) }
			blkDuration    = testcase.LetValue(s, time.Duration(0))
			blkDurationGet = func(t *testcase.T) time.Duration { return blkDuration.Get(t) }
			blkLet         = func(s *testcase.Spec, fn func(*testcase.T, testing.TB)) {
				blkCounterInc := func(t *testcase.T) { blkCounter.Set(t, blkCounter.Get(t)+1) }

				blk.Let(s, func(t *testcase.T) func(assert.It) {
					return func(it assert.It) {
						blkCounterInc(t)
						time.Sleep(blkDurationGet(t))
						fn(t, it)
					}
				})
			}
		)

		s.When(`the assertion fails`, func(s *testcase.Spec) {
			//s.Before(func(t *testcase.T) { t.Skip() }) // TODO

			blkLet(s, func(t *testcase.T, tb testing.TB) { tb.Fail() })

			andMultipleAssertionEventSentToTestingTB := func(s *testcase.Spec) {
				s.And(`and multiple assertion event sent to testing.TB`, func(s *testcase.Spec) {
					cuCounter := testcase.LetValue(s, 0)

					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Cleanup(func() { cuCounter.Set(t, cuCounter.Get(t)+1) })
						tb.Error(`foo`)
						tb.Errorf(`%s`, `baz`)
					})

					stubTB.Let(s, func(t *testcase.T) *doubles.TB {
						stub := &doubles.TB{}
						t.Cleanup(func() {
							t.Must.Contain(stub.Logs.String(), `foo`)
							t.Must.Contain(stub.Logs.String(), `baz`)
						})
						t.Cleanup(stub.Finish)
						return stub
					})

					s.Then(`list events replied to the passed testing.TB`, func(t *testcase.T) {
						act(t)
					})

					s.Then(`cleanup is forwarded regardless the failed error`, func(t *testcase.T) {
						act(t)

						t.Must.True(0 < cuCounter.Get(t))
					})
				})
			}

			s.And(`strategy don't allow further retries other than the first evaluation`, func(s *testcase.Spec) {
				strategyWillRetry.LetValue(s, false)

				s.Then(`it will execute the assertion at least once`, func(t *testcase.T) {
					act(t)

					t.Must.Equal(1, blkCounterGet(t))
				})

				s.Then(`it will fail the test`, func(t *testcase.T) {
					act(t)

					assert.Must(t).True(stubTB.Get(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)
			})

			s.And(`strategy will allow further retries even over the fist assertion block evaluation`, func(s *testcase.Spec) {
				strategyWillRetry.LetValue(s, true)

				s.Then(`it will run for as long as the wait timeout duration`, func(t *testcase.T) {
					act(t)

					assert.Must(t).True(strategyStub.Get(t).IsMaxReached())
				})

				s.Then(`it will execute the condition multiple times`, func(t *testcase.T) {
					act(t)

					assert.Must(t).True(1 < blkCounterGet(t))
				})

				s.Then(`it will fail the test`, func(t *testcase.T) {
					act(t)

					assert.Must(t).True(stubTB.Get(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)

				s.And(`it fails with FailNow`, func(s *testcase.Spec) {
					hasRun := testcase.LetValue(s, false)

					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Cleanup(func() { hasRun.Set(t, true) })
						tb.FailNow()
					})

					s.Then(`it will fail the test`, func(t *testcase.T) {
						sandbox.Run(func() { act(t) })

						assert.Must(t).True(stubTB.Get(t).Failed())
					})

					s.Then(`it will ensure that Cleanup was executed`, func(t *testcase.T) {
						sandbox.Run(func() { act(t) })

						assert.Must(t).True(hasRun.Get(t))
					})
				})
			})
		})

		s.When(`the assertion returns with list happy`, func(s *testcase.Spec) {
			blkLet(s, func(t *testcase.T, tb testing.TB) {
				// nothing to do, TB then will not fail // tb.Success
			})

			andMultipleAssertionEventSentToTestingTB := func(s *testcase.Spec) {
				s.And(`and multiple assertion event sent to testing.TB`, func(s *testcase.Spec) {
					cuCounter := testcase.LetValue(s, 0)

					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Log(`foo`)
						tb.Logf(`%s - %s`, `bar`, `baz`)
						tb.Cleanup(func() { cuCounter.Set(t, cuCounter.Get(t)+1) })
					})

					stubTB.Let(s, func(t *testcase.T) *doubles.TB {
						stub := &doubles.TB{}
						t.Cleanup(stub.Finish)
						t.Cleanup(func() {
							t.Must.Contain(stub.Logs.String(), "foo")
							t.Must.Contain(stub.Logs.String(), "bar - baz")
						})
						return stub
					})

					s.Then(`list events replied to the passed testing.TB`, func(t *testcase.T) {
						act(t)
					})

					s.Then(`cleanup is forwarded`, func(t *testcase.T) {
						act(t)
						stubTB.Get(t).Finish()
						t.Must.True(0 < cuCounter.Get(t))
					})
				})
			}

			s.And(`strategy will not retry the assertion block after the first execution`, func(s *testcase.Spec) {
				strategyWillRetry.LetValue(s, false)

				s.Then(`it will execute the condition at least once`, func(t *testcase.T) {
					act(t)

					t.Must.Equal(1, blkCounterGet(t))
				})

				s.Then(`it will not mark the passed TB as failed`, func(t *testcase.T) {
					act(t)

					assert.Must(t).True(!stubTB.Get(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)
			})

			s.And(`strategy allow multiple condition`, func(s *testcase.Spec) {
				strategyWillRetry.LetValue(s, true)

				s.Then(`it will not use up list the retry strategy loop iterations because the condition doesn't need it`, func(t *testcase.T) {
					act(t)

					assert.Must(t).True(!strategyStub.Get(t).IsMaxReached())
				})

				s.Then(`it will execute the condition only for the required required amount of times`, func(t *testcase.T) {
					act(t)

					t.Must.Equal(1, blkCounterGet(t))
				})

				s.Then(`it will not mark the passed TB as failed`, func(t *testcase.T) {
					act(t)

					assert.Must(t).True(!stubTB.Get(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)

				s.Context(`but it will fail during the Cleanup`, func(s *testcase.Spec) {
					stubTB.Let(s, func(t *testcase.T) *doubles.TB {
						return &doubles.TB{}
					})

					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Cleanup(func() {
							tb.Logf(`I'm running and I'm pointing to %T`, tb)
							tb.FailNow()
						})
					})

					s.Then(`then cleanup is replied to the test subject`, func(t *testcase.T) {
						act(t) // assertion in the TB mock
					})
				})

				s.And(`assertion fails a few times but then yields success`, func(s *testcase.Spec) {
					stubTB.Let(s, func(t *testcase.T) *doubles.TB {
						stub := &doubles.TB{}
						t.Cleanup(stub.Finish)
						t.Cleanup(func() {
							t.Must.False(stub.IsFailed)
						})
						return stub
					})

					var (
						cleanups       = testcase.Let(s, func(t *testcase.T) []string { return []string{} })
						cleanupsAppend = func(t *testcase.T, v ...string) {
							cleanups.Set(t, append(cleanups.Get(t), v...))
						}
					)
					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Cleanup(func() { cleanupsAppend(t, `foo`) })
						tb.Cleanup(func() { cleanupsAppend(t, `bar`) })
						tb.Cleanup(func() { cleanupsAppend(t, `baz`) })

						// fail happens after the cleanups intentionally
						if i := blkCounterGet(t); i < 3 {
							tb.FailNow()
						}
					})

					s.Then(`failed runs cleanup after themselves`, func(t *testcase.T) {
						act(t) // expectations in in the TB input as mock

						expected := []string{
							`baz`, `bar`, `foo`, // block runs first
							`baz`, `bar`, `foo`, // block runs for the second time
						}

						t.Must.Equal(expected, cleanups.Get(t))
					})
				})
			})
		})

		s.When(`the original testing.TB's FailNow is called`, func(s *testcase.Spec) {
			expectedITMessage := let.String(s)
			expectedOuterTFatalMessage := let.String(s)
			blkLet(s, func(t *testcase.T, tb testing.TB) {
				tb.Error(expectedITMessage.Get(t))
				stubTB.Get(t).Fatal(expectedOuterTFatalMessage.Get(t))
			})
			strategyWillRetry.LetValue(s, true)

			s.Then("the assertion won't be retried", func(t *testcase.T) {
				act(t)
				t.Must.True(stubTB.Get(t).Failed())
				t.Must.Equal(1, blkCounter.Get(t))
				t.Must.Contain(stubTB.Get(t).Logs.String(), expectedITMessage.Get(t))
				t.Must.Contain(stubTB.Get(t).Logs.String(), expectedOuterTFatalMessage.Get(t))
			})
		})

		s.When(`the original testing.TB's Fail is called`, func(s *testcase.Spec) {
			expectedITMessage := let.String(s)
			expectedOuterTErrorMessage := let.String(s)
			blkLet(s, func(t *testcase.T, tb testing.TB) {
				tb.Error(expectedITMessage.Get(t))
				stubTB.Get(t).Error(expectedOuterTErrorMessage.Get(t))
			})
			strategyWillRetry.LetValue(s, true)

			s.Then("the assertion won't be retried", func(t *testcase.T) {
				act(t)
				t.Must.True(stubTB.Get(t).Failed())
				t.Must.Equal(1, blkCounter.Get(t))
				t.Must.Contain(stubTB.Get(t).Logs.String(), expectedITMessage.Get(t))
				t.Must.Contain(stubTB.Get(t).Logs.String(), expectedOuterTErrorMessage.Get(t))
				t.Must.Contain(stubTB.Get(t).Logs.String(), "failed during Eventually.Assert")
			})

			s.And("the original testing tb was already failed", func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) { stubTB.Get(t).Fail() })

				s.Then("the assertion retry is acceptable, since it is not necessarily related to our act", func(t *testcase.T) {
					act(t)
					t.Must.True(stubTB.Get(t).Failed())
					t.Must.Equal(42, blkCounter.Get(t))
					t.Must.Contain(stubTB.Get(t).Logs.String(), expectedITMessage.Get(t))
					t.Must.Contain(stubTB.Get(t).Logs.String(), expectedOuterTErrorMessage.Get(t))
				})
			})
		})
	})
}

func TestRetry_Assert_failsOnceButThenPass(t *testing.T) {
	w := assert.Eventually{
		RetryStrategy: assert.Waiter{
			WaitDuration: 0,
			Timeout:      42 * time.Second,
		},
	}

	var (
		stub    = &doubles.TB{}
		counter int
		times   int
	)
	w.Assert(stub, func(it assert.It) {
		// run 42 times
		// 41 times it will fail but create cleanup
		// on the last it should pass
		//
		it.Cleanup(func() { counter++ })
		if 41 <= times {
			return
		}
		times++
		it.Fail()
	})

	t.Log("it is a design decision that the last cleanup is not executed during the assert looping")
	t.Log("the value might be still expected to be used.")
	assert.Must(t).Equal(41, counter)

	stub.Finish()
	assert.Must(t).Equal(42, counter)
}

func TestRetry_Assert_panic(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	w := assert.Eventually{
		RetryStrategy: assert.RetryStrategyFunc(func(condition func() bool) {
			for condition() {
			}
		}),
	}
	expectedPanicValue := rnd.String()
	dtb := &doubles.TB{}
	ro := sandbox.Run(func() {
		w.Assert(dtb, func(it assert.It) {
			panic(expectedPanicValue)
		})
	})
	assert.True(t, ro.Goexit, "expected that dtb called FailNow")
	assert.True(t, dtb.IsFailed)
	assert.Must(t).Contain(dtb.Logs.String(), fmt.Sprintf("panic: %s", expectedPanicValue))
}

type stubRetryStrategy struct {
	ShouldRetry bool
	counter     int
}

func (s *stubRetryStrategy) IsMaxReached() bool {
	return 42 <= s.counter
}

func (s *stubRetryStrategy) inc() bool {
	s.counter++

	return !s.IsMaxReached()
}

func (s *stubRetryStrategy) While(condition func() bool) {
	for condition() && s.inc() && s.ShouldRetry {
	}
}

func TestRetryCount_While(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		i        = testcase.Var[int]{ID: `max times`}
		strategy = testcase.Let(s, func(t *testcase.T) assert.RetryStrategy {
			return assert.RetryCount(i.Get(t))
		})
		condition = testcase.Var[bool]{ID: `condition`}
		subject   = func(t *testcase.T) int {
			var count int
			strategy.Get(t).While(func() bool {
				count++
				return condition.Get(t)
			})
			return count
		}
	)

	s.When(`max times is 0`, func(s *testcase.Spec) {
		i.LetValue(s, 0)

		s.And(`condition always yields true`, func(s *testcase.Spec) {
			condition.LetValue(s, true)

			s.Then(`it should run at least one times`, func(t *testcase.T) {
				t.Must.Equal(1, subject(t))
			})
		})

		s.And(`condition always yields false`, func(s *testcase.Spec) {
			condition.LetValue(s, false)

			s.Then(`it should stop on the first iteration`, func(t *testcase.T) {
				t.Must.Equal(1, subject(t))
			})
		})
	})

	s.When(`max times is greater than 0`, func(s *testcase.Spec) {
		i.Let(s, func(t *testcase.T) int {
			return random.New(random.CryptoSeed{}).IntBetween(1, 10)
		})

		s.And(`condition always yields true`, func(s *testcase.Spec) {
			condition.LetValue(s, true)

			s.Then(`it should run for the maximum retry count plus one for the initial run`, func(t *testcase.T) {
				t.Must.Equal(i.Get(t)+1, subject(t))
			})
		})

		s.And(`condition always yields false`, func(s *testcase.Spec) {
			condition.LetValue(s, false)

			s.Then(`it should stop on the first iteration`, func(t *testcase.T) {
				t.Must.Equal(1, subject(t))
			})
		})
	})
}

func TestEventuallyWithin(t *testing.T) {
	t.Run("time.Duration", func(t *testing.T) {
		t.Run("on timeout", func(t *testing.T) {
			it := assert.MakeIt(t)
			e := assert.EventuallyWithin(time.Millisecond)
			dtb := &doubles.TB{}

			t1 := time.Now()
			e.Assert(dtb, func(it assert.It) { it.Fail() })
			t2 := time.Now()

			it.Must.True(dtb.IsFailed)

			duration := t2.Sub(t1)
			it.Must.True(time.Millisecond <= duration)
		})
		t.Run("within the time", func(t *testing.T) {
			it := assert.MakeIt(t)
			e := assert.EventuallyWithin(time.Millisecond)
			dtb := &doubles.TB{}

			t1 := time.Now()
			e.Assert(dtb, func(it assert.It) {})
			t2 := time.Now()

			it.Must.False(dtb.IsFailed)

			duration := t2.Sub(t1)
			it.Must.True(duration <= time.Millisecond)
		})
	})
	t.Run("retry count", func(t *testing.T) {
		t.Run("out of count", func(t *testing.T) {
			it := assert.MakeIt(t)
			e := assert.EventuallyWithin(3)
			dtb := &doubles.TB{}

			e.Assert(dtb, func(it assert.It) {
				it.Fail()
			})

			it.Must.True(dtb.IsFailed)
		})
		t.Run("within the count", func(t *testing.T) {
			it := assert.MakeIt(t)

			e := assert.EventuallyWithin(3)
			dtb := &doubles.TB{}

			n := 3
			e.Assert(dtb, func(it assert.It) {
				if n == 0 {
					return
				}
				n--
				it.Fail()
			})

			it.Must.False(dtb.IsFailed)
		})
	})
}
