package internal_test

import (
	"context"
	"runtime"
	"sync"
	"testing"

	"github.com/adamluzsi/testcase/internal"
	"github.com/stretchr/testify/require"
)

func TestTeardown_Defer_order(t *testing.T) {
	td := &internal.Teardown{}
	var res []int
	td.Defer(func() { res = append(res, 3) })
	td.Defer(func() { res = append(res, 2) })
	td.Defer(func() { res = append(res, 1) })
	td.Defer(func() { res = append(res, 0) })
	td.Finish()
	//
	require.Equal(t, []int{0, 1, 2, 3}, res)
}

func TestTeardown_Defer_commonFunctionSignatures(t *testing.T) {
	td := &internal.Teardown{}
	var res []int
	td.Defer(func() error { res = append(res, 1); return nil })
	td.Defer(func() { res = append(res, 0) })
	td.Finish()
	//
	require.Equal(t, []int{0, 1}, res)
}

func TestTeardown_Defer_ignoresGoExit(t *testing.T) {
	t.Run(`spike`, func(t *testing.T) {
		var a, b, c bool
		internal.InGoroutine(func() {
			defer func() {
				a = true
			}()
			defer func() {
				b = true
				runtime.Goexit()
			}()
			defer func() {
				c = true
			}()
			runtime.Goexit()
		})
		//
		require.True(t, a)
		require.True(t, b)
		require.True(t, c)
	})

	var a, b, c bool
	internal.InGoroutine(func() {
		td := &internal.Teardown{}
		defer td.Finish()
		td.Defer(func() {
			a = true
		})
		td.Defer(func() {
			b = true
			runtime.Goexit()
		})
		td.Defer(func() {
			c = true
		})
		runtime.Goexit()
	})
	//
	require.True(t, a)
	require.True(t, b)
	require.True(t, c)
}

func TestTeardown_Defer_withinCleanup(t *testing.T) {
	var a, b, c bool
	td := &internal.Teardown{}
	td.Defer(func() {
		a = true
		td.Defer(func() {
			b = true
			td.Defer(func() {
				c = true
			})
		})
	})
	td.Finish()
	//
	require.True(t, a)
	require.True(t, b)
	require.True(t, c)
}

func TestTeardown_Defer_args(t *testing.T) {
	td := &internal.Teardown{}
	t.Run(`arg is primitive type`, func(t *testing.T) {
		fn := func(_ int) {}

		t.Run(`proper input`, func(t *testing.T) {
			require.NotPanics(t, func() { td.Defer(fn, 42) })
		})

		t.Run(`invalid input`, func(t *testing.T) {
			const msg = `deferred function argument[0] type mismatch: expected int, but got string from`
			message := getPanicMessage(t, func() { td.Defer(fn, "42") })
			require.Contains(t, message, msg)
		})
	})

	t.Run(`arg is interface type`, func(t *testing.T) {
		fn := func(ctx context.Context) {}

		t.Run(`proper input`, func(t *testing.T) {
			require.NotPanics(t, func() { td.Defer(fn, context.Background()) })
		})

		t.Run(`invalid input`, func(t *testing.T) {
			const msg = `deferred function argument[0] string doesn't implements context.Context from`
			message := getPanicMessage(t, func() { td.Defer(fn, "42") })
			require.Contains(t, message, msg)
		})
	})

	t.Run(`pass by value`, func(t *testing.T) {
		td := &internal.Teardown{}
		v := 42
		var out int
		td.Defer(func(n int) { out = n }, v)
		v++
		td.Finish()
		require.Equal(t, 42, out)
	})
}

func TestT_Defer_withArgumentsButArgumentCountMismatch(t *testing.T) {
	var subject = func() {
		td := &internal.Teardown{}
		td.Defer(func(text string) {}, `this would be ok`, `but this extra argument is not ok`)
	}

	t.Run(`it will panics early on to help ease the pain of seeing mistakes`, func(t *testing.T) {
		require.Panics(t, func() { subject() })
	})

	t.Run(`panic message will include hint`, func(t *testing.T) {
		message := getPanicMessage(t, func() { subject() })
		require.Contains(t, message, `/Teardown_test.go`)
		require.Contains(t, message, `expected 1`)
		require.Contains(t, message, `got 2`)
	})

	t.Run(`interface type with wrong implementation`, func(t *testing.T) {
		type notContextForSure struct{}
		var fn = func(ctx context.Context) {}
		var subject = func(ctx interface{}) {
			td := &internal.Teardown{}
			td.Defer(fn, ctx)
		}
		require.Panics(t, func() { subject(notContextForSure{}) })
		message := getPanicMessage(t, func() { subject(notContextForSure{}) })
		require.Contains(t, message, `Teardown_test.go`)
		require.Contains(t, message, `doesn't implements context.Context`)
		require.Contains(t, message, `argument[0]`)
	})
}

func TestTeardown_Defer_runtimeGoexit(t *testing.T) {
	t.Run(`spike`, func(t *testing.T) {
		var ran bool
		t.Run(``, func(t *testing.T) {
			t.Cleanup(func() { ran = true })
			t.Cleanup(func() { runtime.Goexit() })
		})
		require.True(t, ran)
	})

	var ran bool
	td := &internal.Teardown{}
	td.Defer(func() { ran = true })
	td.Defer(func() { runtime.Goexit() })
	td.Defer(func() { runtime.Goexit() })
	td.Finish()
	require.True(t, ran)
}

func TestTeardown_Defer_CallerOffset(t *testing.T) {
	var subject = func(offset int) string {
		td := &internal.Teardown{CallerOffset: offset}
		return getPanicMessage(t, func() { offsetHelper(td, func(int) {}, "42") })
	}
	require.Contains(t, subject(0), `offset_helper_test.go:5`)
	require.Contains(t, subject(1), `Teardown_test.go`)
}

func TestTeardown_Defer_isThreadSafe(t *testing.T) {
	var (
		td       = &internal.Teardown{}
		out      = &sync.Map{}
		sampling = runtime.NumCPU() * 42

		start sync.WaitGroup
		wg    sync.WaitGroup
	)

	start.Add(1)
	for i := 0; i < sampling; i++ {
		n := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			start.Wait()
			td.Defer(func() {
				out.Store(n, struct{}{})
			})
		}()
	}
	t.Log(`begin race condition`)
	start.Done() // begin
	t.Log(`wait for all the register to finish`)
	wg.Wait()
	t.Log(`execute registered teardown functions`)
	td.Finish()

	for i := 0; i < sampling; i++ {
		_, ok := out.Load(i)
		require.True(t, ok)
	}
}

func TestTeardown_Finish_idempotent(t *testing.T) {
	var count int
	td := &internal.Teardown{}
	td.Defer(func() { count++ })
	td.Finish()
	td.Finish()
	require.Equal(t, 1, count)
}

func getPanicMessage(tb testing.TB, fn func()) (r string) {
	defer func() {
		var ok bool
		r, ok = recover().(string)
		require.True(tb, ok, `expected to panic`)
	}()
	fn()
	return
}
