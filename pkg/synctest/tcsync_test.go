package synctest_test

import (
	"context"
	"runtime"
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

	p.Release() // wait no longer blocks
}

func TestPhaser(t *testing.T) {
	s := testcase.NewSpec(t)

	phaser := testcase.Let(s, func(t *testcase.T) *synctest.Phaser {
		var p synctest.Phaser
		t.Cleanup(p.Finish)
		return &p
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

		assert.Within(t, time.Millisecond, func(ctx context.Context) {
			<-phaser.Get(t).Done()
		}, "it is expected that phaser no longer blocks on <-Done()")
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
			phaser.Get(t).Release()
		})

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, int32(n), atomic.LoadInt32(&done))
		})

		assert.NotWithin(t, time.Millisecond, func(ctx context.Context) {
			phaser.Get(t).Wait()
		}, "it is expected that phaser is still blocking on wait")

		assert.NotWithin(t, time.Millisecond, func(ctx context.Context) {
			<-phaser.Get(t).Done()
		}, "it is expected that phaser is still blocking on <-Done()")
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
			phaser.Get(t).ReleaseOne()
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

		assert.NotWithin(t, time.Millisecond, func(ctx context.Context) {
			<-phaser.Get(t).Done()
		}, "it is expected that phaser is still blocking on <-Done()")
	})

	s.Test("Release is safe to be called multiple times", func(t *testcase.T) {
		t.Random.Repeat(2, 7, func() {
			phaser.Get(t).Finish()
		})
	})

	s.Test("Done / Wait / chan receive operator", func(t *testcase.T) {
		var c int32 = 2

		go func() {
			defer atomic.AddInt32(&c, -1)
			phaser.Get(t).Wait()
		}()
		go func() {
			defer atomic.AddInt32(&c, -1)
			<-phaser.Get(t).Done()
		}()

		for i := 0; i < 42; i++ {
			runtime.Gosched()
			assert.Equal(t, 2, atomic.LoadInt32(&c))
		}

		phaser.Get(t).Finish()

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, 0, atomic.LoadInt32(&c))
		})
	})

	s.Test("race", func(t *testcase.T) {
		p := phaser.Get(t)

		testcase.Race(func() {
			p.Wait()
		}, func() {
			p.Release()
		}, func() {
			p.ReleaseOne()
		}, func() {
			p.Finish()
		})
	})
}
