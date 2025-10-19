package synctest_test

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/pkg/synctest"
)

func ExamplePhaser() {
	var p synctest.Phaser
	defer p.Finish()

	go func() { p.Wait() }()
	go func() { p.Wait() }()
	go func() { p.Wait() }()

	p.Broadcast() // wait no longer blocks
}

func TestPhaser(t *testing.T) {
	s := testcase.NewSpec(t)

	phaser := testcase.Let(s, func(t *testcase.T) *synctest.Phaser {
		var p synctest.Phaser
		t.Cleanup(p.Finish)
		return &p
	})

	s.Test("smoke #Wait #Finish", func(t *testcase.T) {
		go func() { phaser.Get(t).Wait() }()

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, 1, phaser.Get(t).Len())
		})

		phaser.Get(t).Finish()

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, 0, phaser.Get(t).Len())
		})
	})

	var incJob = func(t *testcase.T, p *synctest.Phaser, c *int32) {
	listening:
		for {
			select {
			case <-t.Done():
				break listening
			default:
				phaser.Get(t).Wait()
				atomic.AddInt32(c, 1)
			}
		}
	}

	s.Test("smoke #Wait #Broadcast", func(t *testcase.T) {
		var (
			p  = phaser.Get(t)
			ns []*int32

			jobCount = t.Random.IntBetween(2, 7)
			sampling = t.Random.IntBetween(1, 7)
		)
		for i := 0; i < jobCount; i++ {
			var n int32
			ptr := &n
			ns = append(ns, ptr)
			go incJob(t, p, ptr)
		}
		for i := 0; i < sampling; i++ {
			t.Eventually(func(t *testcase.T) {
				assert.Equal(t, len(ns), p.Len())
			})

			p.Broadcast() // release all waiter

			t.Eventually(func(t *testcase.T) {
				var total int32
				for _, n := range ns {
					total += atomic.LoadInt32(n)
				}
				assert.Equal(t, total, int32((i+1)*len(ns)))
			})
		}
	})

	s.Test("smoke #Wait #Signal", func(t *testcase.T) {
		var (
			p  = phaser.Get(t)
			ns []*int32

			jobCount = t.Random.IntBetween(2, 7)
			sampling = t.Random.IntBetween(1, 7)
		)
		for i := 0; i < jobCount; i++ {
			var n int32
			ptr := &n
			ns = append(ns, ptr)
			go incJob(t, p, ptr)
		}
		for i := 0; i < sampling; i++ {
			t.Eventually(func(t *testcase.T) {
				assert.Equal(t, len(ns), p.Len())
			})

			p.Signal() // release one waiter

			t.Eventually(func(t *testcase.T) {
				var total int32
				for _, n := range ns {
					total += atomic.LoadInt32(n)
				}
				assert.Equal(t, total, int32(i+1))
			})
		}
	})

	s.Test("wait and release", func(t *testcase.T) {
		var ready, done int32

		n := t.Random.Repeat(1, 7, func() {
			go func() {
				atomic.AddInt32(&ready, 1)
				defer atomic.AddInt32(&done, 1)
				phaser.Get(t).Wait()
			}()
		})

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, int32(n), atomic.LoadInt32(&ready))
		})

		for i := 0; i < 42; i++ {
			runtime.Gosched()
			assert.Equal(t, 0, atomic.LoadInt32(&done))
		}

		assert.Within(t, time.Millisecond, func(ctx context.Context) {
			phaser.Get(t).Finish()
		})

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, int32(n), atomic.LoadInt32(&done))
		})

		assert.Within(t, time.Millisecond, func(ctx context.Context) {
			phaser.Get(t).Wait()
		}, "it is expected that phaser no longer blocks on wait")
	})

	s.Test("wait and broadcast", func(t *testcase.T) {
		var ready, done int32

		n := t.Random.Repeat(1, 7, func() {
			go func() {
				atomic.AddInt32(&ready, 1)
				defer atomic.AddInt32(&done, 1)
				phaser.Get(t).Wait()
			}()
		})

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, int32(n), atomic.LoadInt32(&ready))
		})

		for i := 0; i < 42; i++ {
			runtime.Gosched()
			assert.Equal(t, 0, atomic.LoadInt32(&done))
		}

		assert.Within(t, time.Millisecond, func(ctx context.Context) {
			phaser.Get(t).Broadcast()
		})

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, int32(n), atomic.LoadInt32(&done))
		})

		assert.NotWithin(t, time.Millisecond, func(ctx context.Context) {
			phaser.Get(t).Wait()
		}, "it is expected that phaser is still blocking on wait")
	})

	s.Test("#Finish will act as a permanently continous #Broadcast", func(t *testcase.T) {
		var i int

		t.OnFail(func() {
			t.Logf("i=%d", i)
		})

		var sampling = 3 * runtime.NumCPU()
		for i := 0; i < sampling; i++ {
			var (
				p synctest.Phaser
				c int32

				spam = make(chan struct{})
			)
			go func() {
			work:
				for {
					select {
					case <-t.Done():
						return
					case <-spam:
						break work
					default:
						go func() {
							atomic.AddInt32(&c, 1)
							defer atomic.AddInt32(&c, -1)
							p.Wait()
						}()
					}
				}
			}()

			t.Eventually(func(t *testcase.T) {
				assert.NotEmpty(t, p.Len())
			})

			p.Finish()

			t.Eventually(func(t *testcase.T) {
				assert.Empty(t, p.Len())
			})

			close(spam)

			t.Eventually(func(t *testcase.T) {
				assert.Equal(t, 0, atomic.LoadInt32(&c))
			})

			runtime.GC()
		}
	})

	s.Test("wait and signal", func(t *testcase.T) {
		var ready, done int32

		n := t.Random.Repeat(1, 7, func() {
			go func() {
				atomic.AddInt32(&ready, 1)
				defer atomic.AddInt32(&done, 1)
				phaser.Get(t).Wait()
			}()
		})

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, int32(n), atomic.LoadInt32(&ready))
		})

		for i := 0; i < 42; i++ {
			runtime.Gosched()
			assert.Equal(t, 0, atomic.LoadInt32(&done))
		}

		assert.Within(t, time.Millisecond, func(ctx context.Context) {
			phaser.Get(t).Signal()
		})

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, 1, atomic.LoadInt32(&done))
		})

		t.Random.Repeat(3, 7, func() {
			runtime.Gosched()
			assert.Equal(t, 1, atomic.LoadInt32(&done))
		})

		assert.NotWithin(t, time.Millisecond, func(ctx context.Context) {
			phaser.Get(t).Wait()
		}, "it is expected that phaser is still blocking on wait")
	})

	s.Test("Release is safe to be called multiple times", func(t *testcase.T) {
		t.Random.Repeat(2, 7, func() {
			phaser.Get(t).Finish()
		})
	})

	s.Test("Finish does broadcast", func(t *testcase.T) {
		var (
			p = phaser.Get(t)
			c int32
			n int32
		)

		n = int32(t.Random.Repeat(3, 7, func() {
			go func() {
				atomic.AddInt32(&c, 1)
				defer atomic.AddInt32(&c, -1)
				p.Wait()
			}()
		}))

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, int(n), p.Len())
		}) // eventually all waiter starts to wait

		var sampling = t.Random.IntBetween(32, 128)
		for i := 0; i < sampling; i++ {
			runtime.Gosched()
			assert.Equal(t, n, atomic.LoadInt32(&c),
				"it was expected that none of the waiters finish at this point")
		}

		p.Finish()

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, 0, atomic.LoadInt32(&c))
		})
	})

	s.Test("Wait with Locker", func(t *testcase.T) {
		var m sync.Mutex
		var sl StubLocker

		go func() {
			m.Lock()
			defer m.Unlock()
			phaser.Get(t).Wait(&m, &sl)
		}()

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, phaser.Get(t).Len(), 1)
		})

		phaser.Get(t).Broadcast()

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, sl.UnlockingN(), 1)
			assert.Equal(t, sl.LockingN(), 1)
		})
	})

	s.Test("mixed locker usage", func(t *testcase.T) {
		var (
			p  = phaser.Get(t)
			sl StubLocker
		)
		go func() { p.Wait(&sl) }()
		go func() { p.Wait() }()

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, 2, p.Len())
		})

		p.Broadcast()

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, 0, p.Len())
		})
	})

	s.Test("race", func(t *testcase.T) {
		var (
			p  = phaser.Get(t)
			sl StubLocker
		)
		testcase.Race(func() {
			p.Wait()
		}, func() {
			p.Wait(&sl)
		}, func() {
			p.Broadcast()
		}, func() {
			p.Signal()
		}, func() {
			p.Finish()
		})
	})
}
