package testcase

import (
	"runtime"
	"testing"
)

type recorderTB struct {
	testing.TB
	isFailed bool
	events   []func(testing.TB)
}

func (tb *recorderTB) Replay(oth testing.TB) {
	for _, event := range tb.events {
		event(oth)
	}
}

func (tb *recorderTB) Record(event func(tb testing.TB)) {
	tb.events = append(tb.events, event)
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
