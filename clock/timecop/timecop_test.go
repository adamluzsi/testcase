package timecop_test

import (
	"testing"
	"time"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/clock"
	"go.llib.dev/testcase/clock/timecop"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/random"
	"go.llib.dev/testcase/sandbox"
)

var rnd = random.New(random.CryptoSeed{})

func TestSetSpeed(t *testing.T) {
	t.Run("on zero", func(t *testing.T) {
		dtb := &doubles.TB{}
		defer dtb.Finish()
		sandbox.Run(func() {
			timecop.SetSpeed(dtb, 0)
		})
		assert.True(t, dtb.IsFailed)
	})
	t.Run("on minus value", func(t *testing.T) {
		dtb := &doubles.TB{}
		defer dtb.Finish()
		sandbox.Run(func() {
			timecop.SetSpeed(dtb, -42)
		})
		assert.True(t, dtb.IsFailed)
	})
	t.Run("on positive value", func(t *testing.T) {
		timecop.SetSpeed(t, 10000000)
		s := clock.TimeNow()
		time.Sleep(time.Millisecond)
		e := clock.TimeNow()
		assert.True(t, time.Hour < e.Sub(s))
	})
	t.Run("on frozen time SetSpeed don't start the time", func(t *testing.T) {
		now := time.Now()
		timecop.Travel(t, now, timecop.Freeze())
		timecop.SetSpeed(t, rnd.Float64())
		time.Sleep(time.Microsecond)
		got := clock.TimeNow()
		assert.True(t, now.Equal(got))
	})
}

const buffer = 500 * time.Millisecond

func TestTravel_duration(t *testing.T) {
	t.Run("on no travel", func(t *testing.T) {
		t1 := time.Now()
		t2 := clock.TimeNow()
		assert.True(t, t1.Equal(t2) || t1.Before(t2))
	})
	t.Run("on travel forward", func(t *testing.T) {
		d := time.Duration(rnd.IntB(100, 200)) * time.Second
		timecop.Travel(t, d)
		tnow := time.Now()
		cnow := clock.TimeNow()
		assert.True(t, tnow.Before(cnow))
		assert.True(t, cnow.Sub(tnow) <= d+buffer)
	})
	t.Run("on travel backward", func(t *testing.T) {
		d := time.Duration(rnd.IntB(100, 200)) * time.Second
		timecop.Travel(t, d*-1)
		tnow := time.Now()
		cnow := clock.TimeNow()
		assert.True(t, tnow.Add(d*-1-buffer).Before(cnow))
		assert.True(t, tnow.Add(d*-1+buffer).After(cnow))
	})
}

func TestTravel_timeTime(t *testing.T) {
	t.Run("on no travel", func(t *testing.T) {
		t1 := time.Now()
		t2 := clock.TimeNow()
		assert.True(t, t1.Equal(t2) || t1.Before(t2))
	})
	t.Run("on travel", func(t *testing.T) {
		now := time.Now()
		var (
			year   = rnd.IntB(0, now.Year())
			month  = time.Month(rnd.IntB(1, 12))
			day    = rnd.IntB(1, 20)
			hour   = rnd.IntB(1, 23)
			minute = rnd.IntB(1, 59)
			second = rnd.IntB(1, 59)
			nano   = rnd.IntB(1, int(time.Microsecond-1))
		)
		date := time.Date(year, month, day, hour, minute, second, nano, time.Local)
		timecop.Travel(t, date)
		got := clock.TimeNow()
		assert.Equal(t, time.Local, got.Location())
		assert.Equal(t, year, got.Year())
		assert.Equal(t, month, got.Month())
		assert.Equal(t, day, got.Day())
		assert.Equal(t, hour, got.Hour())
		assert.Equal(t, minute, got.Minute())
		assert.True(t, second-1 <= got.Second() && got.Second() <= second+1)
		assert.True(t, nano-int(buffer) <= got.Nanosecond() && got.Nanosecond() <= nano+int(buffer))
	})
	t.Run("on travel with freeze", func(t *testing.T) {
		now := time.Now()
		var (
			year   = rnd.IntB(0, now.Year())
			month  = time.Month(rnd.IntB(1, 12))
			day    = rnd.IntB(1, 20)
			hour   = rnd.IntB(1, 23)
			minute = rnd.IntB(1, 59)
			second = rnd.IntB(1, 59)
			nano   = rnd.IntB(1, int(time.Microsecond-1))
		)
		date := time.Date(year, month, day, hour, minute, second, nano, time.Local)
		timecop.Travel(t, date, timecop.Freeze())
		time.Sleep(time.Millisecond)
		got := clock.TimeNow()
		assert.True(t, date.Equal(got))
		assert.Waiter{WaitDuration: time.Second}.Wait()
		assert.True(t, date.Equal(clock.TimeNow()))
	})
	t.Run("on travel with freeze, then unfreeze", func(t *testing.T) {
		now := time.Now()
		var (
			year   = rnd.IntB(0, now.Year())
			month  = time.Month(rnd.IntB(1, 12))
			day    = rnd.IntB(1, 20)
			hour   = rnd.IntB(1, 23)
			minute = rnd.IntB(1, 59)
			second = rnd.IntB(1, 59)
			nano   = rnd.IntB(1, int(time.Microsecond-1))
		)
		date := time.Date(year, month, day, hour, minute, second, nano, time.Local)
		timecop.Travel(t, date, timecop.Freeze())
		time.Sleep(time.Millisecond)
		got := clock.TimeNow()
		assert.True(t, date.Equal(got))
		assert.Waiter{WaitDuration: time.Second}.Wait()
		assert.True(t, date.Equal(clock.TimeNow()))
		timecop.Travel(t, clock.TimeNow(), timecop.Unfreeze())
		assert.MakeRetry(time.Second).Assert(t, func(it assert.It) {
			it.Must.False(date.Equal(clock.TimeNow()))
		})
	})
}

func TestTravel_cleanup(t *testing.T) {
	date := time.Now().AddDate(-10, 0, 0)
	t.Run("", func(t *testing.T) {
		timecop.Travel(t, date, timecop.Freeze())
		assert.Equal(t, date.Year(), clock.TimeNow().Year())
	})
	const msg = "was not expected that timecop travel leak out from the sub test"
	assert.NotEqual(t, date.Year(), clock.TimeNow().Year(), msg)
}
