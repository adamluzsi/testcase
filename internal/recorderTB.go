package internal

import (
	"runtime"
	"sync"
	"testing"
)

type RecorderTB struct {
	testing.TB
	IsFailed bool
	Config   struct {
		Passthrough bool
	}

	records []*record
}

type record struct {
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

func (rtb *RecorderTB) Record(blk func(r *record)) {
	rec := &record{}
	blk(rec)
	rtb.records = append(rtb.records, rec)
	rec.play(rtb.Config.Passthrough)
}

func (rtb *RecorderTB) Forward() {
	defer rtb.withPassthrough()()
	for _, record := range rtb.records {
		record.Forward()
	}
}

func (rtb *RecorderTB) CleanupNow() {
	defer rtb.withPassthrough()()
	InGoroutine(func() {
		for _, event := range rtb.records {
			if event.Cleanup != nil {
				defer event.Cleanup()
			}
		}
	})
}

func (rtb *RecorderTB) withPassthrough() func() {
	currentPassthrough := rtb.Config.Passthrough
	rtb.Config.Passthrough = true
	return func() { rtb.Config.Passthrough = currentPassthrough }
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
	rtb.Record(func(r *record) {
		r.Forward = func() { rtb.TB.Cleanup(f) }
		r.Cleanup = f
	})
}

func (rtb *RecorderTB) Helper() {
	rtb.Record(func(r *record) {
		r.Forward = func() { rtb.TB.Helper() }
	})
}

func (rtb *RecorderTB) Log(args ...interface{}) {
	rtb.Record(func(r *record) {
		r.Forward = func() { rtb.TB.Log(args...) }
	})
}

func (rtb *RecorderTB) Logf(format string, args ...interface{}) {
	rtb.Record(func(r *record) {
		r.Forward = func() { rtb.TB.Logf(format, args...) }
	})
}

func (rtb *RecorderTB) markFailed() {
	rtb.IsFailed = true
}

func (rtb *RecorderTB) Fail() {
	rtb.Record(func(r *record) {
		r.Forward = func() { rtb.TB.Fail() }
		r.Ensure = func() { rtb.markFailed() }
	})
}

func (rtb *RecorderTB) failNow() {
	rtb.markFailed()
	runtime.Goexit()
}

func (rtb *RecorderTB) FailNow() {
	rtb.Record(func(r *record) {
		r.Forward = func() { rtb.TB.FailNow() }
		r.Mimic = func() { rtb.failNow() }
		r.Ensure = func() { rtb.markFailed() }
	})
}

func (rtb *RecorderTB) Error(args ...interface{}) {
	rtb.Record(func(r *record) {
		r.Forward = func() { rtb.TB.Error(args...) }
		r.Ensure = func() { rtb.markFailed() }
	})
}

func (rtb *RecorderTB) Errorf(format string, args ...interface{}) {
	rtb.Record(func(r *record) {
		r.Forward = func() { rtb.TB.Errorf(format, args...) }
		r.Ensure = func() { rtb.markFailed() }
	})
}

func (rtb *RecorderTB) Fatal(args ...interface{}) {
	rtb.Record(func(r *record) {
		r.Forward = func() { rtb.TB.Fatal(args...) }
		r.Mimic = func() { rtb.failNow() }
		r.Ensure = func() { rtb.markFailed() }
	})
}

func (rtb *RecorderTB) Fatalf(format string, args ...interface{}) {
	rtb.Record(func(r *record) {
		r.Forward = func() { rtb.TB.Fatalf(format, args...) }
		r.Mimic = func() { rtb.failNow() }
		r.Ensure = func() { rtb.markFailed() }
	})
}

func (rtb *RecorderTB) Failed() bool {
	var failed bool
	rtb.Record(func(r *record) {
		r.Forward = func() { failed = rtb.TB.Failed() }
		r.Mimic = func() { failed = rtb.IsFailed }
	})
	return failed
}
