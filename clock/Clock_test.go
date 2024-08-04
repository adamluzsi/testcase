package clock_test

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/pp"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/clock"
	"go.llib.dev/testcase/clock/internal"
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

	s.Test("deep freezing things before calling After will make the newly made after not moving", func(t *testcase.T) {
		timecop.Travel(t, time.Duration(0), timecop.DeepFreeze)

		var (
			duration      = time.Microsecond
			assertTimeout = time.Millisecond
			afterChannel  = clock.After(duration)
		)

		var tryReadChannel = func(ctx context.Context) {
			select {
			case <-afterChannel:
			case <-ctx.Done():
			}
		}

		assert.NotWithin(t, assertTimeout, tryReadChannel,
			"expected that channel is not readable due to deep freeze")

		timecop.Travel(t, duration/2, timecop.DeepFreeze)
		assert.NotWithin(t, assertTimeout, tryReadChannel,
			"even after travelling a shorter duration than the After(timeout)",
			"it should be still not ticking off")

		timecop.Travel(t, duration/2+time.Nanosecond, timecop.DeepFreeze)
		assert.Within(t, assertTimeout, tryReadChannel,
			"after time travel went to a time where the after should have ended already")

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
	const failureRateMultiplier float64 = 0.70
	var adjust = func(n int64) int64 {
		return int64(float64(n) * failureRateMultiplier)
	}
	s := testcase.NewSpec(t)

	duration := testcase.Let[time.Duration](s, nil)
	ticker := testcase.Let(s, func(t *testcase.T) *clock.Ticker {
		ticker := clock.NewTicker(duration.Get(t))
		t.Defer(ticker.Stop)
		return ticker
	})

	s.Test("by default, clock.Ticker behaves as time.Ticker", func(t *testcase.T) {
		duration.Set(t, time.Second/1000)

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

		time.Sleep(time.Second / 10)
		close(done)
		wg.Wait()

		assert.True(t, 10 < timeTicks)
		assert.True(t, adjust(100/10) <= timeTicks)
		assert.True(t, adjust(100/10) <= clockTicks)
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
				t.Log("ticker ticked", pp.Format(at))
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

	s.Test("freezing will not affect the frequency of the ticks only the returned time, as ticks often used for background scheduling", func(t *testcase.T) {
		timecop.Travel(t, time.Duration(0), timecop.Freeze)
		duration.Set(t, time.Second/10)

		var ticks int64
		go func() {
			for {
				select {
				case <-ticker.Get(t).C:
					atomic.AddInt64(&ticks, 1)
				case <-t.Done():
					return
				}
			}
		}()

		time.Sleep(duration.Get(t) * 2)
		assert.True(t, 0 < atomic.LoadInt64(&ticks))

		const additionalTicks = 10000
		timecop.Travel(t, duration.Get(t)*additionalTicks)
		runtime.Gosched()

		assert.Eventually(t, 2*duration.Get(t), func(t assert.It) {
			currentTicks := atomic.LoadInt64(&ticks)
			expMinTicks := int64(additionalTicks * failureRateMultiplier)
			t.Log("additional ticks:", additionalTicks)
			t.Log("current ticks:", currentTicks)
			t.Log("min exp ticks:", expMinTicks)
			assert.True(t, expMinTicks < currentTicks)
		})
	})

	s.Test("deep freeze that happened before the creation of ticker will make them halted from the get go", func(t *testcase.T) {
		timecop.Travel(t, time.Duration(0), timecop.DeepFreeze)
		duration.Set(t, time.Microsecond)

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

		time.Sleep(time.Second / 10)
		assert.Equal(t, atomic.LoadInt64(&ticks), 0)
	})

	s.Test("deep freeze that happened during the ticker's lifetime will affect the frequency of the ticks as it will make it halt", func(t *testcase.T) {
		duration.Set(t, time.Second)
		timecop.Travel(t, time.Duration(0), timecop.DeepFreeze)
		_, ok := internal.Check()
		assert.True(t, ok)

		var ticks int64
		go func() {
			for {
				select {
				case <-ticker.Get(t).C:
					atomic.AddInt64(&ticks, 1)
				case <-t.Done():
					return
				}
			}
		}()

		// time.Sleep(time.Second)
		// travel 3 tick ahead
		timecop.Travel(t, 3*duration.Get(t)+time.Nanosecond, timecop.DeepFreeze)

		time.Sleep(3 * time.Second)

		assert.Eventually(t, time.Second, func(t assert.It) {
			assert.Equal(t, atomic.LoadInt64(&ticks), 3)
		})
	})

	s.TODO("travelling backwards will make the ticks freeze in time until the last ticked at is reached")

	s.Test("travelling forward with deep freeze flag will cause the ticker to tick the amount it should have if the time was spent", func(t *testcase.T) {
		duration.Set(t, time.Microsecond)

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

		time.Sleep(time.Second / 20)

		timecop.Travel(t, time.Duration(0), timecop.DeepFreeze)
		time.Sleep(2 * duration.Get(t))

		var ticksAfterFreezing int64
		assert.Eventually(t, time.Second, func(t assert.It) {
			curTicks := atomic.LoadInt64(&ticks)
			if ticksAfterFreezing == curTicks {
				return
			}
			ticksAfterFreezing = curTicks
			runtime.Gosched()
			time.Sleep(time.Nanosecond)
		})

		time.Sleep(time.Second / 20)
		assert.True(t, int64(float64(ticksAfterFreezing)*failureRateMultiplier) <= atomic.LoadInt64(&ticks))
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

		time.Sleep(time.Second / 10)
		assert.True(t, adjust(100/10) <= atomic.LoadInt64(&ticks))
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

		time.Sleep(time.Second / 10)
		assert.True(t, adjust(100/10) <= atomic.LoadInt64(&ticks))
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
		time.Sleep(time.Second/10 + time.Microsecond)
		runtime.Gosched()
		var expectedTickCount int64 = adjust(100 / 10)
		t.Log("exp:", expectedTickCount, "got:", atomic.LoadInt64(&ticks))
		assert.True(t, expectedTickCount <= atomic.LoadInt64(&ticks))

		timecop.SetSpeed(t, 1000) // 100x times faster
		time.Sleep(time.Second/10 + time.Microsecond)
		runtime.Gosched()
		expectedTickCount += adjust(100 / 10 * 1000)
		t.Log("exp:", expectedTickCount, "got:", atomic.LoadInt64(&ticks))
		assert.True(t, expectedTickCount <= atomic.LoadInt64(&ticks))
		// *FLAKY
	})

	t.Run("race", func(t *testing.T) {
		ticker := clock.NewTicker(time.Nanosecond)
		const timeout = 50 * time.Millisecond

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
