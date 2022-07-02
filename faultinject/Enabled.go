package faultinject

import (
	"fmt"
	"os"
	"strconv"
	"sync"
)

func init() { initEnabled() }

func initEnabled() {
	state.Enabled = false
	const envKey = "TESTCASE_FAULTINJECT"
	v, ok := os.LookupEnv(envKey)
	if !ok {
		return
	}
	enabled, err := strconv.ParseBool(v)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %s", envKey, err.Error())
		return
	}
	state.Enabled = enabled
}

var state struct {
	Mutex    sync.RWMutex
	Counter  int
	Enabled  bool
	Original bool
}

func Enabled() bool {
	state.Mutex.RLock()
	defer state.Mutex.RUnlock()
	return state.Enabled
}

func Enable() (Disable func()) {
	state.Mutex.Lock()
	defer state.Mutex.Unlock()

	if state.Counter == 0 {
		state.Original = state.Enabled
	}

	state.Counter++
	state.Enabled = true

	return func() {
		state.Mutex.Lock()
		defer state.Mutex.Unlock()
		state.Counter--
		if state.Counter == 0 {
			state.Enabled = state.Original
		}
	}
}

type testingTB interface {
	Cleanup(func())
}

func EnableForTest(tb testingTB) {
	tb.Cleanup(Enable())
}
