package clock_test

import (
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/clock"
	"github.com/adamluzsi/testcase/clock/timecop"
	"testing"
	"time"
)

const bufferDuration = 10 * time.Millisecond

func TestTimeNow(t *testing.T) {
	s := testcase.NewSpec(t, testcase.Flaky(time.Second))
	s.HasSideEffect()

	act := func(t *testcase.T) time.Time {
		return clock.TimeNow()
	}

	s.Test("normally it just returns the current time", func(t *testcase.T) {
		t1 := time.Now()
		t2 := act(t)
		t.Must.True(t1.Equal(t2) || t1.Before(t2))
	})

	s.When("TimeCop is moving in time", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			timecop.Travel(t, time.Hour)
		})

		s.Then("the time  it just returns the current time", func(t *testcase.T) {
			t.Must.True(time.Hour-bufferDuration <= act(t).Sub(time.Now()))
		})
	})
}

func TestSleep(t *testing.T) {
	s := testcase.NewSpec(t, testcase.Flaky(time.Second))
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

	s.When("TimeCop change the flow of time", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			timecop.SetFlowOfTime(t, 2)
		})

		s.Then("the time it just returns the current time", func(t *testcase.T) {
			ed := time.Duration(float64(duration.Get(t)+bufferDuration) * 0.5)
			t.Must.True(act(t) <= ed)
		})
	})
}

func TestAfter(t *testing.T) {
	s := testcase.NewSpec(t, testcase.Flaky(time.Second))
	s.HasSideEffect()

	duration := testcase.Let(s, func(t *testcase.T) time.Duration {
		return time.Duration(t.Random.IntB(25, 100)) * time.Millisecond
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

	s.When("TimeCop change the flow of time", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			timecop.SetFlowOfTime(t, 2)
		})

		alteredDuration := testcase.Let(s, func(t *testcase.T) time.Duration {
			return time.Duration(float64(duration.Get(t)+bufferDuration) * 0.5)
		})

		s.Then("clock.After goes faster", func(t *testcase.T) {
			_, d := act(t)
			t.Must.True(d <= alteredDuration.Get(t))
		})

		s.Test("returned time is relatively calculated to the flow of time", func(t *testcase.T) {
			before := time.Now().Add(alteredDuration.Get(t) - bufferDuration)
			gotTime, _ := act(t)
			t.Must.True(before.Before(gotTime) || before.Equal(gotTime))
		})
	})
}
