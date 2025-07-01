package proxy

import (
	"sync"
	"testing"
	"time"
)

var rwm sync.RWMutex

var _TimeNow = time.Now

// TimeNow is designed to be independent of the clock, allowing it to function autonomously.
// Features that are intended to be controlled by the clock should operate independently of any proxies.
// For instance, generating a random unique value should remain unaffected by time travel.
func TimeNow() time.Time {
	rwm.RLock()
	defer rwm.RUnlock()
	return _TimeNow()
}

func StubTimeNow(tb testing.TB, stub func() time.Time) {
	rwm.Lock()
	defer rwm.Unlock()
	prev := _TimeNow
	tb.Cleanup(func() {
		rwm.Lock()
		defer rwm.Unlock()
		_TimeNow = prev
	})
	_TimeNow = stub
}
