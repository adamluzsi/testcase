package timecop_test

import (
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/clock"
	"github.com/adamluzsi/testcase/clock/timecop"
	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/random"
	"github.com/adamluzsi/testcase/sandbox"
	"testing"
	"time"
)

var rnd = random.New(random.CryptoSeed{})

func TestSetFlowOfTime_invalidMultiplier(t *testing.T) {
	t.Run("on zero", func(t *testing.T) {
		dtb := &doubles.TB{}
		defer dtb.Finish()
		sandbox.Run(func() {
			timecop.SetFlowOfTime(dtb, 0)
		})
		assert.True(t, dtb.IsFailed)
	})
	t.Run("on minus value", func(t *testing.T) {
		dtb := &doubles.TB{}
		defer dtb.Finish()
		sandbox.Run(func() {
			timecop.SetFlowOfTime(dtb, -42)
		})
		assert.True(t, dtb.IsFailed)
	})
}

const buffer = 500 * time.Millisecond

func TestTravel(t *testing.T) {
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

func TestTravelTo(t *testing.T) {
	t.Run("on no travel", func(t *testing.T) {
		t1 := time.Now()
		t2 := clock.TimeNow()
		assert.True(t, t1.Equal(t2) || t1.Before(t2))
	})
	t.Run("on travelling", func(t *testing.T) {
		now := time.Now()
		var (
			year  = now.Year()
			month = now.Month()
			day   = now.Day() + rnd.IntB(1, 3)
		)
		timecop.TravelTo(t, year, month, day)
		got := clock.TimeNow()
		assert.Equal(t, year, got.Year())
		assert.Equal(t, month, got.Month())
		assert.Equal(t, day, got.Day())
	})
}
