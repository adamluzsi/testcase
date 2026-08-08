package tcsync_test

import (
	"runtime"
	"sync/atomic"
	"time"
)

const timeout = time.Second / 8

type StubLocker struct {
	_LockingN, _UnlockingN int32
}

func (stub *StubLocker) LockingN() int32 {
	return atomic.LoadInt32(&stub._LockingN)
}

func (stub *StubLocker) UnlockingN() int32 {
	return atomic.LoadInt32(&stub._UnlockingN)
}

func (stub *StubLocker) Lock() {
	atomic.AddInt32(&stub._LockingN, 1)
}

func (stub *StubLocker) Unlock() {
	atomic.AddInt32(&stub._UnlockingN, 1)
}

func wait() {
	for i := 0; i < 128; i++ {
		runtime.Gosched()
		time.Sleep(time.Nanosecond)
	}
}
