package teardown_test

import (
	"context"
	"runtime"
	"sync"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/teardown"
	"go.llib.dev/testcase/sandbox"
)

func TestTeardown_Defer_order(t *testing.T) {
	td := &teardown.Teardown{}
	var res []int
	td.Defer(func() { res = append(res, 3) })
	td.Defer(func() { res = append(res, 2) })
	td.Defer(func() { res = append(res, 1) })
	td.Defer(func() { res = append(res, 0) })
	td.Finish()
	//
	assert.Must(t).Equal([]int{0, 1, 2, 3}, res)
}

func TestTeardown_Defer_commonFunctionSignatures(t *testing.T) {
	td := &teardown.Teardown{}
	var res []int
	td.Defer(func() error { res = append(res, 1); return nil })
	td.Defer(func() { res = append(res, 0) })
	td.Finish()
	//
	assert.Must(t).Equal([]int{0, 1}, res)
}

func TestTeardown_Defer_ignoresGoExit(t *testing.T) {
	t.Run(`spike`, func(t *testing.T) {
		var a, b, c bool
		sandbox.Run(func() {
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
		assert.Must(t).True(a)
		assert.Must(t).True(b)
		assert.Must(t).True(c)
	})

	var a, b, c bool
	sandbox.Run(func() {
		td := &teardown.Teardown{}
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
	assert.Must(t).True(a)
	assert.Must(t).True(b)
	assert.Must(t).True(c)
}

func TestTeardown_Defer_panic(t *testing.T) {
	defer func() { recover() }()
	var a, b, c bool
	const expectedPanicMessage = `boom`

	td := &teardown.Teardown{}
	td.Defer(func() { a = true })
	td.Defer(func() { b = true; panic(expectedPanicMessage) })
	td.Defer(func() { c = true })

	actualPanicValue := func() (r string) {
		defer func() { r = recover().(string) }()
		td.Finish()
		return ""
	}()
	//
	assert.Must(t).True(a)
	assert.Must(t).True(b)
	assert.Must(t).True(c)
	assert.Must(t).Contains(actualPanicValue, expectedPanicMessage)
}

func TestTeardown_Defer_withinCleanup(t *testing.T) {
	var a, b, c bool
	td := &teardown.Teardown{}
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
	assert.Must(t).True(a)
	assert.Must(t).True(b)
	assert.Must(t).True(c)
}

func TestTeardown_Defer_args(t *testing.T) {
	td := &teardown.Teardown{}
	t.Run(`arg is primitive type`, func(t *testing.T) {
		fn := func(_ int) {}

		t.Run(`proper input`, func(t *testing.T) {
			assert.Must(t).NotPanic(func() { td.Defer(fn, 42) })
		})

		t.Run(`invalid input`, func(t *testing.T) {
			const msg = `deferred function argument[0] type mismatch: expected int, but got string from`
			message := getPanicMessage(t, func() { td.Defer(fn, "42") })
			assert.Must(t).Contains(message, msg)
		})
	})

	t.Run(`arg is interface type`, func(t *testing.T) {
		fn := func(ctx context.Context) {}

		t.Run(`proper input`, func(t *testing.T) {
			assert.Must(t).NotPanic(func() { td.Defer(fn, context.Background()) })
		})

		t.Run(`invalid input`, func(t *testing.T) {
			const msg = `deferred function argument[0] string doesn't implements context.Context from`
			message := getPanicMessage(t, func() { td.Defer(fn, "42") })
			assert.Must(t).Contains(message, msg)
		})
	})

	t.Run(`pass by value`, func(t *testing.T) {
		td := &teardown.Teardown{}
		v := 42
		var out int
		td.Defer(func(n int) { out = n }, v)
		v++
		td.Finish()
		assert.Must(t).Equal(42, out)
	})
}

func TestTeardown_Defer_withVariadicArgument(t *testing.T) {
	var total int
	s := testcase.NewSpec(t)
	s.Test("", func(t *testcase.T) {
		t.Defer(func(n int, text ...string) { total++ }, 42)
		t.Defer(func(n int, text ...string) { total++ }, 42, "a")
		t.Defer(func(n int, text ...string) { total++ }, 42, "a", "b")
		t.Defer(func(n int, text ...string) { total++ }, 42, "a", "b", "c")
	})
	s.Finish()
	assert.Must(t).Equal(4, total)
}

func TestTeardown_Defer_withVariadicArgument_argumentPassed(t *testing.T) {
	var total int
	sum := func(v int, ns ...int) {
		total += v
		for _, n := range ns {
			total += n
		}
	}
	s := testcase.NewSpec(t)
	s.Test("", func(t *testcase.T) {
		t.Defer(sum, 1)
		t.Defer(sum, 2, 3)
		t.Defer(sum, 4, 5, 6)
	})
	s.Finish()
	assert.Must(t).Equal(1+2+3+4+5+6, total)
}

func TestT_Defer_withArgumentsButArgumentCountMismatch(t *testing.T) {
	var subject = func() {
		td := &teardown.Teardown{}
		td.Defer(func(text string) {}, `this would be ok`, `but this extra argument is not ok`)
	}

	t.Run(`it will panics early on to help ease the pain of seeing mistakes`, func(t *testing.T) {
		assert.Must(t).Panic(func() { subject() })
	})

	t.Run(`panic message will include hint`, func(t *testing.T) {
		message := getPanicMessage(t, func() { subject() })
		assert.Must(t).Contains(message, `/Teardown_test.go`)
		assert.Must(t).Contains(message, `expected 1`)
		assert.Must(t).Contains(message, `got 2`)
	})

	t.Run(`interface type with wrong implementation`, func(t *testing.T) {
		type notContextForSure struct{}
		var fn = func(ctx context.Context) {}
		var subject = func(ctx interface{}) {
			td := &teardown.Teardown{}
			td.Defer(fn, ctx)
		}
		assert.Must(t).Panic(func() { subject(notContextForSure{}) })
		message := getPanicMessage(t, func() { subject(notContextForSure{}) })
		assert.Must(t).Contains(message, `Teardown_test.go`)
		assert.Must(t).Contains(message, `doesn't implements context.Context`)
		assert.Must(t).Contains(message, `argument[0]`)
	})
}

func TestTeardown_Defer_runtimeGoexit(t *testing.T) {
	t.Run(`spike`, func(t *testing.T) {
		var ran bool
		defer func() { assert.Must(t).True(ran) }()
		t.Run(``, func(t *testing.T) {
			t.Cleanup(func() { ran = true })
			t.Cleanup(func() { runtime.Goexit() })
		})
	})

	sandbox.Run(func() {
		var ran bool
		defer func() { assert.Must(t).True(ran) }()
		td := &teardown.Teardown{}
		td.Defer(func() { ran = true })
		td.Defer(func() { runtime.Goexit() })
		td.Finish()
		assert.Must(t).True(ran)
	})

}

func TestTeardown_Defer_CallerOffset(t *testing.T) {
	var subject = func(offset int) string {
		td := &teardown.Teardown{CallerOffset: offset}
		return getPanicMessage(t, func() { offsetHelper(td, func(int) {}, "42") })
	}
	assert.Must(t).Contains(subject(0), `offset_helper_test.go:7`)
	assert.Must(t).Contains(subject(1), `Teardown_test.go`)
}

func TestTeardown_Defer_isThreadSafe(t *testing.T) {
	var (
		td       = &teardown.Teardown{}
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
		assert.Must(t).True(ok)
	}
}

func TestTeardown_Finish_idempotent(t *testing.T) {
	var count int
	td := &teardown.Teardown{}
	td.Defer(func() { count++ })
	td.Finish()
	td.Finish()
	assert.Must(t).Equal(1, count)
}

func getPanicMessage(tb testing.TB, fn func()) (r string) {
	defer func() {
		var ok bool
		r, ok = recover().(string)
		assert.Must(tb).True(ok, `expected to panic`)
	}()
	fn()
	return
}
