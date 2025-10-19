package synctest

import (
	"sync"
	"sync/atomic"
)

// Phaser is a synchronization primitive for coordinating multiple goroutines.
// I that combines the behavior of a latch, barrier, and phaser.
// Goroutines may register via Wait, be released one at a time via Signal or all at once via Broadcast.
// Ultimately, using Finish can ensure, that all wait is released, to ensure that all waiter is released.
//
// Like sync.Mutex, a zero value Phaser is ready to use right away.
// If you’re using a sync.Mutex to protect shared resources,
// you can pass it as an optional argument to Phaser.Wait.
// This lets the mutex be released while waiting and automatically reacquired upon the end of waiting.
type Phaser struct {
	m sync.Mutex
	o sync.Once
	c *sync.Cond

	len  int64
	done int32
}

type phaserLockerUnlocker func()

func (fn phaserLockerUnlocker) Lock() {}

func (fn phaserLockerUnlocker) Unlock() { fn() }

func (p *Phaser) init() {
	p.o.Do(func() { p.c = sync.NewCond((*nopLocker)(nil)) })
}

func (p *Phaser) Len() int {
	return int(atomic.LoadInt64(&p.len))
}

func (p *Phaser) Wait(ls ...sync.Locker) {
	if atomic.LoadInt32(&p.done) != 0 {
		return
	}

	p.init()

	var ml = multiLocker(ls)

	p.m.Lock()

	if atomic.LoadInt32(&p.done) != 0 {
		p.m.Unlock()
		return
	}

	p.c.L = phaserLockerUnlocker(func() {
		// we increment here, because by this time on locker#Unlock,
		// the sync.Cond's runtime_notifyListAdd is already executed,
		// and listens to Broadcast
		atomic.AddInt64(&p.len, 1)
		ml.Unlock()
		p.c.L = (*nopLocker)(nil) // restore no operation locker
		p.m.Unlock()
	})
	// during sync.Cond#Wait, ml#Unlock will be unlocked,
	// and we need to re-acquire it after the wait finished
	defer ml.Lock()
	// during Wait the len is incremented,
	// and afterwards we need to drecrement it.
	defer atomic.AddInt64(&p.len, -1)
	p.c.Wait()
}

func (p *Phaser) Signal() {
	p.init()
	p.c.Signal()
}

func (p *Phaser) Broadcast() {
	p.init()
	p.c.Broadcast()
}

// Finish lets all waiting goroutines continue immediately.
// After it’s called, any new calls to Wait will also return right away.
func (p *Phaser) Finish() {
	p.m.Lock()
	defer p.m.Unlock()
	p.init()
	atomic.CompareAndSwapInt32(&p.done, 0, 1)
	p.c.Broadcast()
}
