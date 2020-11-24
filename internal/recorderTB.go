package internal

import (
	"runtime"
	"testing"
)

type RecorderTB struct {
	testing.TB
	IsFailed bool
	events   []*recorderTBEvent
}

type recorderTBEvent struct {
	Action    func(testing.TB)
	isCleanup bool
}

func (tb *RecorderTB) Record(action func(tb testing.TB)) *recorderTBEvent {
	event := &recorderTBEvent{Action: action}
	tb.events = append(tb.events, event)
	return event
}

func (tb *RecorderTB) Replay(oth testing.TB) {
	for _, event := range tb.events {
		event.Action(oth)
	}
}

func (tb *RecorderTB) ReplayCleanup(oth testing.TB) {
	for _, event := range tb.events {
		if event.isCleanup {
			event.Action(oth)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (tb *RecorderTB) Cleanup(f func()) {
	tb.Record(func(tb testing.TB) { tb.Cleanup(f) }).isCleanup = true
}

func (tb *RecorderTB) Helper() {
	tb.Record(func(tb testing.TB) { tb.Helper() })
}

func (tb *RecorderTB) Log(args ...interface{}) {
	tb.Record(func(tb testing.TB) { tb.Log(args...) })
}

func (tb *RecorderTB) Logf(format string, args ...interface{}) {
	tb.Record(func(tb testing.TB) { tb.Logf(format, args...) })
}

func (tb *RecorderTB) fail() {
	tb.IsFailed = true
}

func (tb *RecorderTB) Fail() {
	tb.Record(func(tb testing.TB) { tb.Fail() })
	tb.fail()
}

func (tb *RecorderTB) failNow() {
	tb.fail()
	runtime.Goexit()
}

func (tb *RecorderTB) FailNow() {
	tb.Record(func(tb testing.TB) { tb.FailNow() })
	tb.failNow()
}

func (tb *RecorderTB) Error(args ...interface{}) {
	tb.Record(func(tb testing.TB) { tb.Error(args...) })
	tb.fail()
}

func (tb *RecorderTB) Errorf(format string, args ...interface{}) {
	tb.Record(func(tb testing.TB) { tb.Errorf(format, args...) })
	tb.fail()
}

func (tb *RecorderTB) Fatal(args ...interface{}) {
	tb.Record(func(tb testing.TB) { tb.Fatal(args...) })
	tb.failNow()
}

func (tb *RecorderTB) Fatalf(format string, args ...interface{}) {
	tb.Record(func(tb testing.TB) { tb.Fatalf(format, args...) })
	tb.failNow()
}

func (tb *RecorderTB) Failed() bool {
	tb.Record(func(tb testing.TB) { _ = tb.Failed() })

	if tb.TB != nil {
		return tb.TB.Failed()
	}

	return tb.IsFailed
}
