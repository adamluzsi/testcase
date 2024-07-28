package clock_test

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.llib.dev/testcase/let"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/clock"
	"go.llib.dev/testcase/clock/timecop"
)

const BufferTime = 50 * time.Millisecond

func ExampleNow() {
	now := clock.Now()
	_ = now
}

func TestNow(t *testing.T) {
	s := testcase.NewSpec(t)
	s.HasSideEffect()

	act := func(*testcase.T) time.Time {
		return clock.Now()
	}

	s.Test("By default, it just returns the current time", func(t *testcase.T) {
		timeNow := time.Now()
		clockNow := act(t)
		t.Must.True(timeNow.Add(-1 * BufferTime).Before(clockNow))
	}, testcase.Flaky(time.Second))

	s.When("Timecop is moving in time", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			timecop.Travel(t, time.Hour)
		})

		s.Then("the time  it just returns the current time", func(t *testcase.T) {
			t.Must.True(time.Hour-BufferTime <= time.Until(act(t)))
		})

		s.Then("time is still moving forward", func(t *testcase.T) {
			now := act(t)

			t.Eventually(func(it *testcase.T) {
				next := act(t)
				it.Must.False(now.Equal(next))
				it.Must.True(next.After(now))
			})
		})

		s.And("time moved to a specific time given in non-local format", func(s *testcase.Spec) {
			expTime := let.Time(s)

			s.Before(func(t *testcase.T) {
				timecop.Travel(t, expTime.Get(t).UTC(), timecop.Freeze)
			})

			s.Then("the time it just returned in the same Local as time.Now()", func(t *testcase.T) {
				t.Must.Equal(time.Now().Location(), act(t).Location())
			})

			s.Then("the time is what Travel set", func(t *testcase.T) {
				t.Must.True(expTime.Get(t).Equal(act(t)))
			})
		})
	})

	s.When("Timecop is altering the flow of time", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			timecop.SetSpeed(t, 2000)
		})

		s.Then("the speed of time is affected", func(t *testcase.T) {
			start := act(t)
			time.Sleep(time.Millisecond)
			after := act(t)
			duration := after.Sub(start)
			t.Must.True(duration > time.Second)
		})
	})
}

func TestSleep(t *testing.T) {
	s := testcase.NewSpec(t)
	s.HasSideEffect()

	duration := testcase.Let(s, func(t *testcase.T) time.Duration {
		return time.Duration(t.Random.IntB(25, 100)) * time.Millisecond
	})
	act := func(t *testcase.T) time.Duration {
		before := time.Now()
		clock.Sleep(duration.Get(t))
		after := time.Now()
		return after.Sub(before)
	}

	s.Test("By default, it just sleeps as time.Sleep()", func(t *testcase.T) {
		t.Must.True(act(t) <= duration.Get(t)+BufferTime)
	})

	s.When("Timecop change the flow of time", func(s *testcase.Spec) {
		multi := testcase.LetValue[float64](s, 1000)

		s.Before(func(t *testcase.T) {
			timecop.SetSpeed(t, multi.Get(t))
		})

		s.Then("the time it takes to sleep is affected", func(t *testcase.T) {
			expectedMaximumDuration := time.Duration(float64(duration.Get(t))/multi.Get(t)) + BufferTime
			sleptFor := act(t)
			t.Log("expectedMaximumDuration:", expectedMaximumDuration.String())
			t.Log("sleptFor:", sleptFor.String())
			t.Must.True(sleptFor <= expectedMaximumDuration)
		})
	})

	s.Test("timecop travels during sleep", func(t *testcase.T) {
		duration.Set(t, time.Hour)
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			act(t)
			cancel()
		}()

		timecop.Travel(t, time.Second)
		select {
		case <-ctx.Done():
			t.Fatal("was not expected to finish already")
		case <-time.After(BufferTime):
			// OK
		}

		timecop.Travel(t, duration.Get(t))
		select {
		case <-ctx.Done():
			// OK
		case <-time.After(BufferTime):
			t.Fatal("was expected to finish already")
		}
	})
}

func TestAfter(t *testing.T) {
	s := testcase.NewSpec(t)
	s.HasSideEffect()

	duration := testcase.Let(s, func(t *testcase.T) time.Duration {
		return time.Duration(t.Random.IntB(24, 42)) * time.Millisecond
	})
	act := func(t *testcase.T, ctx context.Context) {
		select {
		case <-clock.After(duration.Get(t)):
		case <-ctx.Done(): // assertion already finished
		case <-t.Done(): // test already finished
		}
	}

	buftime := testcase.Let(s, func(t *testcase.T) time.Duration {
		return time.Duration(float64(duration.Get(t)) * 0.2)
	})

	s.Test("By default, it behaves as time.After()", func(t *testcase.T) {
		assert.NotWithin(t, duration.Get(t)-buftime.Get(t), func(ctx context.Context) {
			act(t, ctx)
		})
		assert.Within(t, duration.Get(t)+buftime.Get(t), func(ctx context.Context) {
			act(t, ctx)
		})
	})

	s.When("Timecop change the flow of time's speed", func(s *testcase.Spec) {
		speed := testcase.LetValue[float64](s, 2)

		s.Before(func(t *testcase.T) {
			timecop.SetSpeed(t, speed.Get(t))
		})

		alteredDuration := testcase.Let(s, func(t *testcase.T) time.Duration {
			return time.Duration(float64(duration.Get(t))/speed.Get(t)) + BufferTime
		})

		s.Then("clock.After goes faster", func(t *testcase.T) {
			assert.Within(t, alteredDuration.Get(t)+time.Millisecond, func(ctx context.Context) {
				act(t, ctx)
			})
		})
	})

	s.Test("when time travel happens during waiting on the result of clock.After, then it will affect them.", func(t *testcase.T) {
		duration := time.Hour
		ch := clock.After(duration)

		t.Log("before any travelling, just a regular check if the time is done")
		select {
		case <-ch:
			t.Fatal("it was not expected that clock.After finished already")
		default:
			// OK
		}

		t.Log("travel takes us to a time where the original duration is not yet reached")

		timecop.Travel(t, duration/2)

		select {
		case <-ch:
			t.Fatal("it was not expected that clock.After is already done since we moved less forward than the total duration")
		default:
			// OK
		}

		t.Log("travel takes us after the original duration already reached")
		timecop.Travel(t, (duration/2)+BufferTime)

		select {
		case <-ch:
			// OK
		case <-time.After(3 * time.Second):
			t.Fatal("clock.After should have finished already its work after travel that went more forward as the duration")
		}

	}) //Ω, testcase.Flaky(5*time.Second))

	s.Test("no matter what, when the wait time is zero, clock.After returns instantly", func(t *testcase.T) {
		timecop.SetSpeed(t, 0.001)
		timecop.Travel(t, time.Second, timecop.Freeze)
		assert.Within(t, time.Millisecond, func(ctx context.Context) {
			<-clock.After(0)
		}, "expected to finish instantly")
	})
}

func Test_testTimeWithMinusDuration(t *testing.T) {
	select {
	case <-time.After(-42):
		// OK
	case <-time.After(time.Second):
		t.Fatal("time after was not expected to spend too much time on a minus value")
	}
}

func Test_race(t *testing.T) {
	write := func() {
		timecop.Travel(t, time.Millisecond)
		timecop.SetSpeed(t, 1)
	}
	read := func() {
		clock.Now()
		clock.Sleep(time.Millisecond)
		clock.After(time.Millisecond)
	}
	testcase.Race(write, read, read, read, read)
}

func TestNewTicker(t *testing.T) {
	const failureRateMultiplier = 0.80
	s := testcase.NewSpec(t)

	duration := testcase.Let[time.Duration](s, nil)
	ticker := testcase.Let(s, func(t *testcase.T) *clock.Ticker {
		ticker := clock.NewTicker(duration.Get(t))
		t.Defer(ticker.Stop)
		return ticker
	})

	s.Test("by default, clock.Ticker behaves as time.Ticker", func(t *testcase.T) {
		duration.Set(t, time.Second/100)

		var (
			clockTicks, timeTicks int64
			wg                    sync.WaitGroup
			done                  = make(chan struct{})
		)

		var (
			dur         = duration.Get(t)
			clockTicker = clock.NewTicker(dur)
			timeTicker  = time.NewTicker(dur)
		)
		t.Defer(clockTicker.Stop)
		t.Defer(timeTicker.Stop)

		wg.Add(2)
		go testcase.Race(
			func() {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					case <-clockTicker.C:
						atomic.AddInt64(&clockTicks, 1)
					}
				}
			},
			func() {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					case <-timeTicker.C:
						atomic.AddInt64(&timeTicks, 1)
					}
				}
			},
		)

		time.Sleep(time.Second / 4)
		close(done)
		wg.Wait()

		assert.True(t, 10 < timeTicks)
		assert.True(t, 100/4*failureRateMultiplier <= timeTicks)
		assert.True(t, 100/4*failureRateMultiplier <= clockTicks)
	})

	s.Test("time travelling affect ticks", func(t *testcase.T) {
		duration.Set(t, time.Duration(t.Random.IntBetween(int(time.Minute), int(time.Hour))))
		var (
			now  int64
			done = make(chan struct{})
		)
		defer close(done)
		go func() {
			select {
			case at := <-ticker.Get(t).C:
				t.Log("ticker ticked")
				atomic.AddInt64(&now, at.Unix())
			case <-done:
				return
			}
		}()

		t.Log("normal scenario, no tick yet expected")
		time.Sleep(time.Millisecond)
		runtime.Gosched()
		assert.Empty(t, atomic.LoadInt64(&now), "no tick expected yet")

		t.Log("time travel to the future, but before the tick is suppose to happen")
		timecop.Travel(t, duration.Get(t)/2) // travel to a time in the future where the ticker is still not fired
		runtime.Gosched()
		assert.Empty(t, atomic.LoadInt64(&now), "tick is still not expected")

		t.Log("time travel after the ")
		beforeTravel := clock.Now()
		timecop.Travel(t, (duration.Get(t)/2)+time.Nanosecond) // travel to a time, where the ticker should fire
		runtime.Gosched()

		assert.Eventually(t, time.Second, func(t assert.It) {
			got := atomic.LoadInt64(&now)
			t.Must.NotEmpty(got)
			t.Must.True(beforeTravel.Unix() <= got, "tick is expected at this point")
		})
	})

	s.Test("ticks are continous", func(t *testcase.T) {
		duration.Set(t, time.Second/100)

		var (
			ticks int64
			done  = make(chan struct{})
		)
		defer close(done)
		go func() {
			for {
				select {
				case <-ticker.Get(t).C:
					atomic.AddInt64(&ticks, 1)
				case <-done:
					return
				}
			}
		}()

		time.Sleep(time.Second / 2)
		assert.True(t, 100/2*failureRateMultiplier <= atomic.LoadInt64(&ticks))
	})

	s.Test("duration is scaled", func(t *testcase.T) {
		timecop.SetSpeed(t, 100) // 100x times faster
		duration.Set(t, time.Second)

		var (
			ticks int64
			done  = make(chan struct{})
		)
		defer close(done)
		go func() {
			for {
				select {
				case <-ticker.Get(t).C:
					atomic.AddInt64(&ticks, 1)
				case <-done:
					return
				}
			}
		}()

		time.Sleep(time.Second / 4)
		assert.True(t, 100/4*failureRateMultiplier <= atomic.LoadInt64(&ticks))
	})

	s.Test("duration is scaled midflight", func(t *testcase.T) {
		duration.Set(t, time.Second/100)

		var (
			ticks int64
			done  = make(chan struct{})
		)
		defer close(done)
		go func() {
			for {
				select {
				case <-ticker.Get(t).C:
					atomic.AddInt64(&ticks, 1)
				case <-done:
					return
				}
			}
		}()

		t.Log("ticks:", atomic.LoadInt64(&ticks))
		time.Sleep(time.Second/4 + time.Microsecond)
		runtime.Gosched()
		var expectedTickCount int64 = 100 / 4 * failureRateMultiplier
		t.Log("exp:", expectedTickCount, "got:", atomic.LoadInt64(&ticks))
		assert.True(t, expectedTickCount <= atomic.LoadInt64(&ticks))

		timecop.SetSpeed(t, 1000) // 100x times faster
		time.Sleep(time.Second/4 + time.Microsecond)
		runtime.Gosched()

		// TODO: flaky assertion
		//
		// FLAKY*
		expectedTickCount += 100 / 4 * 1000 * failureRateMultiplier
		t.Log("exp:", expectedTickCount, "got:", atomic.LoadInt64(&ticks))
		assert.True(t, expectedTickCount <= atomic.LoadInt64(&ticks))
		// *FLAKY
	})

	t.Run("race", func(t *testing.T) {
		ticker := clock.NewTicker(time.Minute)
		const timeout = 100 * time.Millisecond

		testcase.Race(
			func() {
				select {
				case <-ticker.C:
				case <-clock.After(timeout):
				}
			},
			func() {
				ticker.Reset(time.Minute)
			},
			func() {
				<-clock.After(timeout)
				ticker.Stop()
			},
		)
	})
}
