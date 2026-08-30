package internal

import (
	"sync"
)

var mutex sync.RWMutex

func mSync() func() {
	mutex.Lock()
	return mutex.Unlock
}

func mRSync() func() {
	mutex.RLock()
	return mutex.RUnlock
}
