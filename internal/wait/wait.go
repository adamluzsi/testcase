package wait

import (
	"runtime"
	"slices"
	"time"
)

func Others(timeout time.Duration) {
	const WaitUnit = time.Nanosecond
	var (
		goroutNum = runtime.NumGoroutine()
		startedAt = time.Now()
	)
	for i := 0; i < goroutNum; i++ { // since goroutines don't have guarantee when they will be scheduled
		runtime.Gosched() // we explicitly mark that we are okay with other goroutines to be scheduled
		elapsed := time.Since(startedAt)
		if timeout <= elapsed { // if max wait time is reached
			return
		}
		if elapsed < timeout { // if we withint the max wait time,
			time.Sleep(WaitUnit) // then we could just yield CPU too with sleep
		}
	}
}

var _ = initScales()
var (
	scales      = map[time.Duration]float64{}
	gscale      float64
	minDuration time.Duration
)

var steps = []time.Duration{
	time.Nanosecond,
	time.Microsecond,
	50 * time.Microsecond,
	100 * time.Microsecond,
	time.Millisecond,
}

func initScales() struct{} {
	slices.Sort(steps)
	var total float64
	for _, d := range steps {
		scale := scaleByDuration(d)
		scales[d] = scale
		total += scale
	}
	gscale = total / float64(len(steps))
	minDuration = time.Duration(scales[time.Nanosecond])
	return struct{}{}
}

func scaleByDuration(d time.Duration) float64 {
	var (
		total time.Duration
		count int = 100
	)
	for i := 0; i < count; i++ {
		start := time.Now()
		time.Sleep(d)
		duration := time.Since(start)
		total += duration
	}
	var avg = float64(total) / float64(count)
	var scale = float64(d) / avg
	return scale
}

func adjust(d time.Duration) time.Duration {
	var s float64 = gscale
	for i := len(steps) - 1; 0 <= i; i-- {
		unit := steps[i]
		if unit < d {
			break
		}
		s = scales[unit]
	}
	return scale(d, s)
}

func scale(d time.Duration, s float64) time.Duration {
	return time.Duration(float64(d) * s)
}

func For(duration time.Duration) {
	if duration <= minDuration {
		return
	}
	runtime.Gosched()
	duration = adjust(duration)
	time.Sleep(duration)
}
