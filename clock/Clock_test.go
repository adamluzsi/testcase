package clock_test

import (
	"context"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/clock"
	"github.com/adamluzsi/testcase/clock/timecop"
)

const bufferDuration = 50 * time.Millisecond

func TestTimeNow(t *testing.T) {
	s := testcase.NewSpec(t)
	s.HasSideEffect()

	act := func(t *testcase.T) time.Time {
		return clock.TimeNow()
	}

	s.Test("normally it just returns the current time", func(t *testcase.T) {
		timeNow := time.Now()
		clockNow := act(t)
		t.Must.True(timeNow.Add(-1 * bufferDuration).Before(clockNow))
	}, testcase.Flaky(time.Second))

	s.When("Timecop is moving in time", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			timecop.Travel(t, time.Hour)
		})

		s.Then("the time  it just returns the current time", func(t *testcase.T) {
			t.Must.True(time.Hour-bufferDuration <= act(t).Sub(time.Now()))
		})

		s.Then("time is still moving forward", func(t *testcase.T) {
			now := act(t)

			t.Eventually(func(it assert.It) {
				next := act(t)
				it.Must.False(now.Equal(next))
				it.Must.True(next.After(now))
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

	s.Test("normally it just sleeps normally", func(t *testcase.T) {
		t.Must.True(act(t) <= duration.Get(t)+bufferDuration)
	})

	s.When("Timecop change the flow of time", func(s *testcase.Spec) {
		multi := testcase.LetValue[float64](s, 1000)

		s.Before(func(t *testcase.T) {
			timecop.SetSpeed(t, multi.Get(t))
		})

		s.Then("the time it just returns the current time", func(t *testcase.T) {
			expectedMaximumDuration := time.Duration(float64(duration.Get(t))/multi.Get(t)) + bufferDuration
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
		case <-time.After(bufferDuration):
			// OK
		}

		timecop.Travel(t, duration.Get(t))
		select {
		case <-ctx.Done():
			// OK
		case <-time.After(bufferDuration):
			t.Fatal("was expected to finish already")
		}
	})
}

func TestAfter(t *testing.T) {
	s := testcase.NewSpec(t)
	s.HasSideEffect()

	duration := testcase.Let(s, func(t *testcase.T) time.Duration {
		return time.Duration(t.Random.IntB(24, 42)) * time.Microsecond
	})
	act := func(t *testcase.T) (time.Time, time.Duration) {
		before := time.Now()
		out := <-clock.After(duration.Get(t))
		after := time.Now()
		return out, after.Sub(before)
	}

	s.Test("normally it just sleeps normally", func(t *testcase.T) {
		before := time.Now()
		gotTime, gotDuration := act(t)
		t.Must.True(gotDuration <= duration.Get(t)+bufferDuration)
		t.Must.True(before.Before(gotTime))
	})

	s.When("Timecop change the flow of time", func(s *testcase.Spec) {
		speed := testcase.LetValue[float64](s, 2)

		s.Before(func(t *testcase.T) {
			timecop.SetSpeed(t, speed.Get(t))
		})

		alteredDuration := testcase.Let(s, func(t *testcase.T) time.Duration {
			return time.Duration(float64(duration.Get(t))/speed.Get(t)) + bufferDuration
		})

		s.Then("clock.After goes faster", func(t *testcase.T) {
			_, d := act(t)
			t.Must.True(d <= alteredDuration.Get(t))
		})

		s.Test("returned time is relatively calculated to the flow of time", func(t *testcase.T) {
			before := time.Now().Add(alteredDuration.Get(t))
			gotTime, _ := act(t)
			t.Must.True(before.Add(-1 * bufferDuration).Before(gotTime))
		})
	})

	s.When("Timecop travel in time", func(s *testcase.Spec) {
		date := testcase.Let(s, func(t *testcase.T) time.Time {
			return t.Random.Time()
		})
		s.Before(func(t *testcase.T) {
			timecop.Travel(t, date.Get(t))
		})

		s.Then("returned time will represent this", func(t *testcase.T) {
			finishedAt, _ := act(t)

			t.Must.True(date.Get(t).Before(finishedAt))
			t.Must.True(date.Get(t).Add(duration.Get(t) + bufferDuration).After(finishedAt))
		})
	})

	s.Test("when time travel is done during clock.After", func(t *testcase.T) {
		duration := time.Hour
		ch := clock.After(duration)

		timecop.Travel(t, time.Second)
		select {
		case <-ch:
			t.Fatal("it was not expected that clock.After is already done since we moved less forward than the total duration")
		default:
			// OK
		}

		timecop.Travel(t, duration+bufferDuration)

		select {
		case <-ch:
			// OK
		case <-time.After(time.Second):
			t.Fatal("clock.After should have finished already its work after travel that went more forward as the duration")
		}
	}) //Ω, testcase.Flaky(5*time.Second))
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
		clock.TimeNow()
		clock.Sleep(time.Millisecond)
		clock.After(time.Millisecond)
	}
	testcase.Race(write, read, read, read, read)
}
