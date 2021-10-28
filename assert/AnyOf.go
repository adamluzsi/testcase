package assert

import (
	"sync"
	"testing"

	"github.com/adamluzsi/testcase/internal"
	"github.com/adamluzsi/testcase/internal/fmterror"
)

type AnyOf struct {
	TB testing.TB
	Fn func(...interface{})

	mutex  sync.Mutex
	passed bool
}

func (ao *AnyOf) Test(blk func(it It)) {
	ao.TB.Helper()
	if ao.isPassed() {
		return
	}
	recorder := &internal.RecorderTB{TB: ao.TB}
	defer recorder.CleanupNow()
	internal.RecoverExceptGoexit(func() {
		ao.TB.Helper()
		blk(makeIt(recorder))
	})
	if recorder.IsFailed {
		return
	}
	ao.mutex.Lock()
	defer ao.mutex.Unlock()
	ao.passed = true
}

func (ao *AnyOf) Finish(msg ...interface{}) {
	ao.TB.Helper()
	if ao.isPassed() {
		return
	}
	ao.Fn(fmterror.Message{
		Method:      "AnyOf",
		Cause:       "None of the .Test succeeded",
		Values:      nil,
		UserMessage: msg,
	})
}

func (ao *AnyOf) isPassed() bool {
	ao.mutex.Lock()
	defer ao.mutex.Unlock()
	return ao.passed
}
