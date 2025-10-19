package synctest

import (
	"sync"
	"sync/atomic"
)

type Phaser struct {
	o sync.Once
	c *sync.Cond
	d int32
}

func (p *Phaser) init() {
	p.o.Do(func() {
		p.c = sync.NewCond((*nopLocker)(nil))
	})
}

func (p *Phaser) Wait() {
	p.init()
	if atomic.LoadInt32(&p.d) == 0 {
		p.c.Wait()
	}
}

func (p *Phaser) Done() <-chan struct{} {
	p.init()
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		p.Wait()
	}()
	return ch
}

func (p *Phaser) ReleaseOne() {
	p.init()
	p.c.Signal()
}

func (p *Phaser) Release() {
	p.init()
	p.c.Broadcast()
}

// Finish lets all waiting goroutines continue immediately.
// After it’s called, any new calls to Wait will also return right away.
func (p *Phaser) Finish() {
	p.init()
	if atomic.CompareAndSwapInt32(&p.d, 0, 1) {
		p.c.Broadcast()
	}
}

type nopLocker struct{}

func (*nopLocker) Lock() {}

func (*nopLocker) Unlock() {}
