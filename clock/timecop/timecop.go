package timecop

import (
	"testing"
	"time"

	"go.llib.dev/testcase/clock/internal"
)

func Travel[D time.Duration | time.Time](tb testing.TB, d D, tos ...TravelOption) {
	tb.Helper()
	guardAgainstParallel(tb)
	opt := toOption(tos)
	switch d := any(d).(type) {
	case time.Duration:
		travelByDuration(tb, d, opt)
	case time.Time:
		travelByTime(tb, d, opt)
	}
}

const BlazingFast = 100

func SetSpeed(tb testing.TB, multiplier float64) {
	tb.Helper()
	guardAgainstParallel(tb)
	if multiplier <= 0 {
		tb.Fatal("Timecop.SetSpeed can't receive zero or negative value")
	}
	tb.Cleanup(internal.SetSpeed(multiplier))
}

// guardAgainstParallel
// is a hack that ensures that there was no testing.T.Parallel() used in the test.
func guardAgainstParallel(tb testing.TB) {
	tb.Helper()
	const key, value = `TEST_CASE_TIMECOP_IN_USE`, "TRUE"
	tb.Setenv(key, value) // will fail on parallel execution
}

func travelByDuration(tb testing.TB, d time.Duration, opt internal.Option) {
	tb.Helper()
	travelByTime(tb, internal.TimeNow().Add(d), opt)
}

func travelByTime(tb testing.TB, target time.Time, opt internal.Option) {
	tb.Helper()
	tb.Cleanup(internal.SetTime(target, opt))
}

// Freeze is a Travel TravelOption, and it instruct travel to freeze the time wherever it lands after the travelling..
const Freeze = freeze(0)

type freeze int

func (freeze) configure(o *internal.Option) {
	o.Freeze = true
}

// DeepFreeze is a Travel TravelOption, and it instruct travel to freeze the time wherever it lands after the travelling.
// It is a stronger level of freezing, and will force tickers and timers to also halt immedietly.
const DeepFreeze = deepFreeze(1)

type deepFreeze int

func (deepFreeze) configure(o *internal.Option) {
	o.Freeze = true
	o.Deep = true
}

// Unfreeze is a Travel TravelOption, and it instruct travel that after the time travelling, the flow of time should continue.
const Unfreeze = unfreeze(0)

type unfreeze int

func (unfreeze) configure(o *internal.Option) {
	o.Unfreeze = true
}
