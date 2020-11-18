package testcase

import (
	"runtime"
	"testing"
)

type recorderTB struct {
	testing.TB
	isFailed bool
	events   []*recorderTBEvent
}

type recorderTBEvent struct {
	Action    func(testing.TB)
	isCleanup bool
}

func (tb *recorderTB) Record(action func(tb testing.TB)) *recorderTBEvent {
	event := &recorderTBEvent{Action: action}
	tb.events = append(tb.events, event)
	return event
}

func (tb *recorderTB) Replay(oth testing.TB) {
	for _, event := range tb.events {
		event.Action(oth)
	}
}

func (tb *recorderTB) ReplayCleanup(oth testing.TB) {
	for _, event := range tb.events {
		if event.isCleanup {
			event.Action(oth)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (tb *recorderTB) Cleanup(f func()) {
	tb.Record(func(tb testing.TB) { tb.Cleanup(f) }).isCleanup = true
}

func (tb *recorderTB) Helper() {
	tb.Record(func(tb testing.TB) { tb.Helper() })
}

func (tb *recorderTB) Log(args ...interface{}) {
	tb.Record(func(tb testing.TB) { tb.Log(args...) })
}

func (tb *recorderTB) Logf(format string, args ...interface{}) {
	tb.Record(func(tb testing.TB) { tb.Logf(format, args...) })
}

func (tb *recorderTB) fail() {
	tb.isFailed = true
}

func (tb *recorderTB) Fail() {
	tb.Record(func(tb testing.TB) { tb.Fail() })
	tb.fail()
}

func (tb *recorderTB) failNow() {
	tb.fail()
	runtime.Goexit()
}

func (tb *recorderTB) FailNow() {
	tb.Record(func(tb testing.TB) { tb.FailNow() })
	tb.failNow()
}

func (tb *recorderTB) Error(args ...interface{}) {
	tb.Record(func(tb testing.TB) { tb.Error(args...) })
	tb.fail()
}

func (tb *recorderTB) Errorf(format string, args ...interface{}) {
	tb.Record(func(tb testing.TB) { tb.Errorf(format, args...) })
	tb.fail()
}

func (tb *recorderTB) Fatal(args ...interface{}) {
	tb.Record(func(tb testing.TB) { tb.Fatal(args...) })
	tb.failNow()
}

func (tb *recorderTB) Fatalf(format string, args ...interface{}) {
	tb.Record(func(tb testing.TB) { tb.Fatalf(format, args...) })
	tb.failNow()
}

func (tb *recorderTB) Failed() bool {
	tb.Record(func(tb testing.TB) { _ = tb.Failed() })

	if tb.TB != nil {
		return tb.TB.Failed()
	}

	return tb.isFailed
}
