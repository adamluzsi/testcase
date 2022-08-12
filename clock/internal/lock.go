package internal

import (
	"sync"
)

var mutex sync.RWMutex

func lock() func() {
	mutex.Lock()
	return mutex.Unlock
}

func rlock() func() {
	mutex.RLock()
	return mutex.RUnlock
}
