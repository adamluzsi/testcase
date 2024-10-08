package doubles

import (
	"runtime"
	"sync"
	"testing"

	"go.llib.dev/testcase/internal/env"

	"go.llib.dev/testcase/internal/teardown"
)

type RecorderTB struct {
	testing.TB
	IsFailed bool
	Config   struct {
		Passthrough bool
	}

	// records might be written concurrently, but it is not expected to receive reads during concurrent writes.
	// That is considered a mistake in the testing suite.
	_records []*record
	m        sync.Mutex
}

type record struct {
	Skip    bool
	Forward func()
	Mimic   func()
	Ensure  func()
	Cleanup func()
}

func (r record) play(passthrough bool) {
	if r.Ensure != nil {
		r.Ensure()
	}
	if passthrough {
		r.Forward()
	} else if r.Mimic != nil {
		r.Mimic()
	}
}

func (rtb *RecorderTB) records() []*record {
	rtb.m.Lock()
	defer rtb.m.Unlock()
	var out []*record = make([]*record, len(rtb._records))
	copy(out, rtb._records)
	return out
}

func (rtb *RecorderTB) record(blk func(r *record)) {
	rtb.m.Lock()
	defer rtb.m.Unlock()
	rec := &record{}
	blk(rec)
	rtb._records = append(rtb._records, rec)
	rec.play(rtb.Config.Passthrough)
}

func (rtb *RecorderTB) Forward() {
	rtb.TB.Helper()
	// set passthrough for future events like Recorder used from a .Cleanup callback.
	_ = rtb.withPassthrough()
	for _, record := range rtb.records() {
		if !record.Skip {
			record.Forward()
		}
	}
}

func (rtb *RecorderTB) CleanupNow() {
	rtb.TB.Helper()
	defer rtb.withPassthrough()()
	td := &teardown.Teardown{}
	for _, event := range rtb.records() {
		if event.Cleanup != nil && !event.Skip {
			td.Defer(event.Cleanup)
			event.Skip = true
		}
	}
	td.Finish()
}

func (rtb *RecorderTB) withPassthrough() func() {
	rtb.m.Lock()
	defer rtb.m.Unlock()
	currentPassthrough := rtb.Config.Passthrough
	rtb.Config.Passthrough = true
	return func() {
		rtb.m.Lock()
		defer rtb.m.Unlock()
		rtb.Config.Passthrough = currentPassthrough
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (rtb *RecorderTB) Run(_ string, blk func(testing.TB)) bool {
	sub := &RecorderTB{TB: rtb}
	defer sub.CleanupNow()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		blk(sub)
	}()
	wg.Wait()

	if sub.IsFailed {
		rtb.IsFailed = true

		if rtb.Config.Passthrough {
			rtb.TB.Fail()
		}
	}
	return !sub.IsFailed
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (rtb *RecorderTB) Cleanup(f func()) {
	rtb.record(func(r *record) {
		r.Forward = func() {
			rtb.TB.Helper()
			rtb.TB.Cleanup(f)
		}
		r.Cleanup = f
	})
}

func (rtb *RecorderTB) Helper() {
	rtb.record(func(r *record) {
		r.Forward = func() { rtb.TB.Helper() }
	})
}

func (rtb *RecorderTB) Log(args ...interface{}) {
	rtb.record(func(r *record) {
		r.Forward = func() {
			rtb.TB.Helper()
			rtb.TB.Log(args...)
		}
	})
}

func (rtb *RecorderTB) Logf(format string, args ...interface{}) {
	rtb.record(func(r *record) {
		r.Forward = func() {
			rtb.TB.Helper()
			rtb.TB.Logf(format, args...)
		}
	})
}

func (rtb *RecorderTB) markFailed() {
	rtb.IsFailed = true
}

func (rtb *RecorderTB) Fail() {
	rtb.record(func(r *record) {
		r.Forward = func() {
			rtb.TB.Helper()
			rtb.TB.Fail()
		}
		r.Ensure = func() { rtb.markFailed() }
	})
}

func (rtb *RecorderTB) failNow() {
	rtb.TB.Helper()
	rtb.markFailed()
	runtime.Goexit()
}

func (rtb *RecorderTB) FailNow() {
	rtb.record(func(r *record) {
		r.Forward = func() {
			rtb.TB.Helper()
			rtb.TB.FailNow()
		}
		r.Mimic = func() {
			rtb.TB.Helper()
			rtb.failNow()
		}
		r.Ensure = func() { rtb.markFailed() }
	})
}

func (rtb *RecorderTB) Error(args ...interface{}) {
	rtb.record(func(r *record) {
		r.Forward = func() {
			rtb.TB.Helper()
			rtb.TB.Error(args...)
		}
		r.Ensure = func() { rtb.markFailed() }
	})
}

func (rtb *RecorderTB) Errorf(format string, args ...interface{}) {
	rtb.record(func(r *record) {
		r.Forward = func() {
			rtb.TB.Helper()
			rtb.TB.Errorf(format, args...)
		}
		r.Ensure = func() { rtb.markFailed() }
	})
}

func (rtb *RecorderTB) Fatal(args ...interface{}) {
	rtb.record(func(r *record) {
		r.Forward = func() {
			rtb.TB.Helper()
			rtb.TB.Fatal(args...)
		}
		r.Mimic = func() {
			rtb.TB.Helper()
			rtb.failNow()
		}
		r.Ensure = func() { rtb.markFailed() }
	})
}

func (rtb *RecorderTB) Fatalf(format string, args ...interface{}) {
	rtb.record(func(r *record) {
		r.Forward = func() {
			rtb.TB.Helper()
			rtb.TB.Fatalf(format, args...)
		}
		r.Mimic = func() {
			rtb.TB.Helper()
			rtb.failNow()
		}
		r.Ensure = func() { rtb.markFailed() }
	})
}

func (rtb *RecorderTB) Failed() bool {
	var failed bool
	rtb.record(func(r *record) {
		r.Forward = func() {
			rtb.TB.Helper()
			failed = rtb.TB.Failed()
		}
		r.Mimic = func() {
			rtb.TB.Helper()
			failed = rtb.IsFailed
		}
	})
	return failed
}

func (rtb *RecorderTB) Setenv(key, value string) {
	env.SetEnv(rtb, key, value)
}
