package faultinject

import (
	"os"
	"strings"
	"sync"
)

func init() { initEnabled() }

func initEnabled() {
	enabled.State = false
	const envKey = "TESTCASE_FAULT_INJECTION"
	if v, ok := os.LookupEnv(envKey); ok {
		switch strings.ToUpper(v) {
		case "TRUE", "ON":
			enabled.State = true
		case "FALSE", "OFF":
			enabled.State = false
		}
	}
}

var enabled struct {
	Mutex    sync.Mutex
	Counter  int
	State    bool
	Original bool
}

func Enabled() bool {
	enabled.Mutex.Lock()
	defer enabled.Mutex.Unlock()
	return enabled.State
}

func Enable() (Disable func()) {
	enabled.Mutex.Lock()
	defer enabled.Mutex.Unlock()

	if enabled.Counter == 0 {
		enabled.Original = enabled.State
	}

	enabled.Counter++
	enabled.State = true

	return func() {
		enabled.Mutex.Lock()
		defer enabled.Mutex.Unlock()
		enabled.Counter--
		if enabled.Counter == 0 {
			enabled.State = enabled.Original
		}
	}
}

type testingTB interface {
	Cleanup(func())
}

func EnableForTest(tb testingTB) {
	tb.Cleanup(Enable())
}
