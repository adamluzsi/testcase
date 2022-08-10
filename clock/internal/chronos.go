package internal

import "time"

var Chronos struct {
	Offset     time.Duration
	FlowOfTime float64
}

func init() {
	Chronos.FlowOfTime = 1
}

func DurationFor(d time.Duration) time.Duration {
	return time.Duration(float64(d) / Chronos.FlowOfTime)
}
