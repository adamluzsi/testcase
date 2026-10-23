package testcase_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/contracts"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/internal/environ"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/sandbox"

	"go.llib.dev/testcase/random"

	"go.llib.dev/testcase"
)

var _ testing.TB = &testcase.T{}

func TestT_implementsTestingTB(t *testing.T) {
	testcase.RunSuite(t, contracts.TestingTB{
		Subject: func(t *testcase.T) testing.TB {
			stub := &doubles.TB{}
			t.Cleanup(stub.Finish)
			return testcase.NewTWithSpec(stub, nil)
		},
	})
}

func TestVar_Set_canBeUsedDuringTest(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Context(`runtime define`, func(s *testcase.Spec) {
		nog := testcase.Let(s, func(t *testcase.T) int { return rand.Intn(42) })
		mog := testcase.Let(s, func(t *testcase.T) int { return rand.Intn(42) + 100 })

		var exampleMultiReturnFunc = func(t *testcase.T) (int, int) {
			return nog.Get(t), mog.Get(t)
		}

		s.Context(`Let being set during testCase runtime`, func(s *testcase.Spec) {
			n := testcase.Var[int]{ID: "n"}
			m := testcase.Var[int]{ID: "m"}

			s.Before(func(t *testcase.T) {
				nv, mv := exampleMultiReturnFunc(t)
				n.Set(t, nv)
				m.Set(t, mv)
			})

			s.Test(`let values which are defined during runtime present in the testCase`, func(t *testcase.T) {
				assert.Must(t).Equal(n.Get(t), nog.Get(t))
				assert.Must(t).Equal(m.Get(t), mog.Get(t))
			})
		})
	})

	s.Context(`runtime update`, func(s *testcase.Spec) {
		var initValue = rand.Intn(42)
		x := testcase.Let(s, func(t *testcase.T) int { return initValue })

		s.Before(func(t *testcase.T) {
			x.Set(t, x.Get(t)+1)
		})

		s.Before(func(t *testcase.T) {
			x.Set(t, x.Get(t)+1)
		})

		s.Test(`let will returns the value then override the runtime vars`, func(t *testcase.T) {
			assert.Must(t).Equal(initValue+2, x.Get(t))
		})
	})

}

func TestT_Defer(t *testing.T) {
	var res []int

	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)

		s.Context(``, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				res = append(res, 0)
			})

			s.After(func(t *testcase.T) {
				res = append(res, -1)
			})

			s.Context(``, func(s *testcase.Spec) {
				s.Around(func(t *testcase.T) func() {
					res = append(res, 1)
					return func() { res = append(res, -2) }
				})

				s.Context(``, func(s *testcase.Spec) {
					wDefer := testcase.Let(s, func(t *testcase.T) int {
						t.Defer(func() { res = append(res, -3) })
						return 42
					})

					s.Before(func(t *testcase.T) {
						// calling a variable that has defer will ensure
						// that the deferred function call will be executed
						// as part of the *T#defer stack, and not afterwards
						assert.Must(t).Equal(42, wDefer.Get(t))
					})

					s.Test(``, func(t *testcase.T) {
						t.Defer(func() { res = append(res, -4) })
					})
				})
			})
		})
	})

	assert.Must(t).Equal([]int{0, 1, -4, -3, -2, -1}, res)
}

// TB#Cleanup https://github.com/golang/go/issues/41355
//
//goland:noinspection GoDeferGo
func TestT_Defer_failNowWillNotHang(t *testing.T) {
	assert.Within(t, time.Second, func(ctx context.Context) {
		sandbox.Run(func() {
			s := testcase.NewSpec(&doubles.TB{})
			s.Test(``, func(t *testcase.T) {
				t.Defer(func() { t.FailNow() })

				panic(`die`)
			})
		})
	})
}

func TestT_Defer_whenItIsCalledDuringTestBlock(t *testing.T) {
	var itRan bool
	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.Test(``, func(t *testcase.T) { t.Defer(func() { itRan = true }) })
	})
	assert.Must(t).True(itRan, `then it is expected to ran`)
}

func TestT_Defer_withArguments(t *testing.T) {
	var (
		expected = rand.Int() + 1
		actually int
	)

	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)
		type S struct{ ID int }

		v := testcase.Let(s, func(t *testcase.T) *S {
			s := &S{ID: expected}
			t.Defer(func(id int) { actually = id }, s.ID)
			return s
		})

		s.Test(`testCase that alter the content of value`, func(t *testcase.T) {
			v.Get(t).ID = 0
		})

		s.Test(`interface type with concrete input must be allowed`, func(t *testcase.T) {
			var fn = func(ctx context.Context) {}
			t.Defer(fn, context.Background())
		})
	})

	assert.Must(t).Equal(expected, actually)
}

func TestT_Defer_runsOnlyAfterTestIsdone(t *testing.T) {
	s := testcase.NewSpec(t)

	CTX := testcase.Let(s, func(t *testcase.T) func() context.Context {
		return func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			t.Defer(cancel)
			return ctx
		}
	})

	s.Before(func(t *testcase.T) {
		t.Cleanup(func() { assert.Must(t).NoError(CTX.Get(t)().Err()) })
		t.Defer(func() { assert.Must(t).NoError(CTX.Get(t)().Err()) })
	})

	s.Test("", func(t *testcase.T) {})
}

func TestT_Defer_withArgumentsButArgumentCountMismatch(t *testing.T) {
	s := testcase.NewSpec(t)

	var getPanicMessage = func(fn func()) (r string) {
		defer func() { r, _ = recover().(string) }()
		fn()
		return
	}

	v := testcase.Let(s, func(t *testcase.T) int {
		t.Defer(func(text string) {}, `this would be ok`, `but this extra argument is not ok`)
		return 42
	})

	s.Test(`testCase that it will panics early on to help ease the pain of seeing mistakes`, func(t *testcase.T) {
		assert.Must(t).Panic(func() { _ = v.Get(t) })
	})

	s.Test(`panic message`, func(t *testcase.T) {
		message := getPanicMessage(func() { _ = v.Get(t) })
		assert.Must(t).Contains(message, `/testcase/T_test.go`)
		assert.Must(t).Contains(message, `expected 1`)
		assert.Must(t).Contains(message, `got 2`)
	})

	s.Test(`interface type with wrong implementation`, func(t *testcase.T) {
		type notContextForSure struct{}
		var fn = func(ctx context.Context) {}
		assert.Must(t).Panic(func() { t.Defer(fn, notContextForSure{}) })
		message := getPanicMessage(func() { t.Defer(fn, notContextForSure{}) })
		assert.Must(t).Contains(message, `/testcase/T_test.go`)
		assert.Must(t).Contains(message, `doesn't implements context.Context`)
		assert.Must(t).Contains(message, `argument[0]`)
	})
}

func TestT_Defer_withArgumentsButArgumentTypeMismatch(t *testing.T) {
	s := testcase.NewSpec(t)

	v := testcase.Let(s, func(t *testcase.T) int {
		t.Defer(func(n int) {}, `this is not ok`)
		return 42
	})

	s.Test(`testCase that it will panics early on to help ease the pain of seeing mistakes`, func(t *testcase.T) {
		assert.Must(t).Panic(func() { _ = v.Get(t) })
	})

	s.Test(`panic message`, func(t *testcase.T) {
		message := func() (r string) {
			defer func() { r = recover().(string) }()
			_ = v.Get(t)
			return ``
		}()

		assert.Must(t).Contains(message, `/testcase/T_test.go`)
		assert.Must(t).Contains(message, `expected int`)
		assert.Must(t).Contains(message, `got string`)
	})
}

func TestT_TB(t *testing.T) {
	s := testcase.NewSpec(t)

	for i := 0; i < 10; i++ {
		var ts []testing.TB
		s.Test(`*testcase.TB is set to the given testcase's *testing.T`, func(t *testcase.T) {
			assert.Must(t).NotNil(t.TB)
			assert.Must(t).NotContains(ts, t.TB, `TB should be unique for each testCase run`)
			ts = append(ts, t.TB)
		})
	}
}

func TestT_Defer_calledWithoutFunctionAndWillPanic(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Test(`defer expected to panic for non function input as first value`, func(t *testcase.T) {
		var withReturnValue = func() int { return 42 }

		assert.Must(t).Panic(func() { t.Defer(withReturnValue()) })
	})

	s.Test(`defer expected to panic for invalid inputs`, func(t *testcase.T) {
		var dummyClose = func() error { return nil }
		pv := assert.Must(t).Panic(func() { t.Defer(dummyClose()) })
		assert.Must(t).Contains(pv, `T#Defer can only take functions`)
	})

}

func TestT_Defer_willRunEvenIfSomethingForceTheTestToStopEarly(t *testing.T) {
	var ran bool
	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.Before(func(t *testcase.T) { t.Defer(func() { ran = true }) })
		s.Test(``, func(t *testcase.T) { t.Skip(`please stop early`) })
	})
	assert.Must(t).True(ran)
}

func TestT_HasTag(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Context(`a`, func(s *testcase.Spec) {
		s.Tag(`a`)

		s.Context(`b`, func(s *testcase.Spec) {
			s.Tag(`b`)

			s.Context(`c`, func(s *testcase.Spec) {
				s.Tag(`c`)

				s.Test(`c`, func(t *testcase.T) {
					assert.Must(t).True(t.HasTag(`a`))
					assert.Must(t).True(t.HasTag(`b`))
					assert.Must(t).True(t.HasTag(`c`))
					assert.Must(t).True(!t.HasTag(`d`))
				})
			})

			s.Test(`b`, func(t *testcase.T) {
				assert.Must(t).True(t.HasTag(`a`))
				assert.Must(t).True(t.HasTag(`b`))
				assert.Must(t).True(!t.HasTag(`c`))
				assert.Must(t).True(!t.HasTag(`d`))
			})
		})

		s.Test(`a`, func(t *testcase.T) {
			assert.Must(t).True(t.HasTag(`a`))
			assert.Must(t).True(!t.HasTag(`b`))
			assert.Must(t).True(!t.HasTag(`c`))
			assert.Must(t).True(!t.HasTag(`d`))
		})
	})

	s.Test(``, func(t *testcase.T) {
		assert.Must(t).True(!t.HasTag(`a`))
		assert.Must(t).True(!t.HasTag(`b`))
		assert.Must(t).True(!t.HasTag(`c`))
		assert.Must(t).True(!t.HasTag(`d`))
	})
}

func TestT_Random(t *testing.T) {
	randomGenerationWorks := func(t *testcase.T) {
		assert.Retry{Strategy: assert.Waiter{WaitDuration: time.Second}}.Assert(t, func(it testing.TB) {
			assert.True(it, 0 < t.Random.Int())
		})
	}

	t.Run(`when environment value is set`, func(t *testing.T) {
		testcase.SetEnv(t, environ.KeySeed, `42`)
		s := testcase.NewSpec(t)
		s.Test(``, func(t *testcase.T) {
			assert.Must(t).NotEmpty(t.Random)

			randomGenerationWorks(t)
		})
	})

	s := testcase.NewSpec(t)
	s.Test(``, func(t *testcase.T) {
		randomGenerationWorks(t)
	})
}

func TestT_Eventually(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})

	t.Run(`with default eventually retry strategy`, func(t *testing.T) {
		stub := &doubles.TB{}
		s := testcase.NewSpec(stub)
		s.HasSideEffect()
		var eventuallyRan bool
		s.Test(``, func(t *testcase.T) {
			t.Eventually(func(it *testcase.T) {
				eventuallyRan = true
				assert.Must(it).True(t.Random.Bool())
			}) // eventually pass
		})
		stub.Finish()
		s.Finish()
		assert.Must(t).True(!stub.IsFailed, `expected to pass`)
		assert.Must(t).True(eventuallyRan)
	})

	t.Run(`with config passed`, func(t *testing.T) {
		stub := &doubles.TB{}
		var strategyUsed bool
		strategy := assert.LoopFunc(func(condition func() bool) {
			strategyUsed = true
			for condition() {
			}
		})
		s := testcase.NewSpec(stub, testcase.WithRetryStrategy(strategy))
		s.HasSideEffect()
		s.Test(``, func(t *testcase.T) {
			t.Eventually(func(it *testcase.T) {
				assert.Must(it).True(t.Random.Bool())
			}) // eventually pass
		})
		stub.Finish()
		s.Finish()
		assert.Must(t).True(!stub.IsFailed, `expected to pass`)
		assert.Must(t).True(strategyUsed, `retry strategy of the eventually call was used`)
	})

	t.Run("Eventually uses a testcase.T that allows its functionalities all from the the eventually block", func(t *testing.T) {
		// After extensive testing, we discovered that constantly switching between using `assert.It` and `testcase.T` is risky
		// and prone to errors without adding any benefit.
		// Therefore, our next step is to streamline the API to simplify testing.

		stub := &doubles.TB{}
		s := testcase.NewSpec(stub)
		s.HasSideEffect()

		expTag := rnd.StringNC(5, random.CharsetAlpha())
		s.Tag(expTag)

		expVal := rnd.Int()
		v := testcase.LetValue(s, expVal)

		var ran bool
		s.Test(``, func(tcT *testcase.T) {
			tcT.Eventually(func(it *testcase.T) {
				ran = true
				assert.Equal(t, expVal, v.Get(it))
				assert.Equal(t, tcT.Random, it.Random)
				assert.True(t, tcT.HasTag(expTag))
				assert.True(t, tcT.TB != it.TB)
			})
		})

		stub.Finish()
		s.Finish()

		assert.Must(t).True(!stub.IsFailed, `expected to pass`)
		assert.True(t, ran)
	})

	t.Run("smoke", func(t *testing.T) {
		stub := &doubles.TB{}
		s := testcase.NewSpec(stub)
		s.HasSideEffect()

		var ran bool
		s.Test(``, func(tcT *testcase.T) {
			var failed bool
			tcT.Eventually(func(t *testcase.T) {
				if !failed {
					failed = true
					t.FailNow()
				}
				// OK
				ran = true
			})
			assert.False(t, tcT.Failed())
		})

		stub.Finish()
		s.Finish()

		assert.Must(t).True(!stub.IsFailed, `expected to pass`)
		assert.True(t, ran)
	})

	t.Run("when failure occurs during the variable initialisation", func(t *testing.T) {
		t.Run("permanently", func(t *testing.T) {
			stub := &doubles.TB{}
			s := testcase.NewSpec(stub, testcase.WithRetryStrategy(assert.RetryCount(3)))
			s.HasSideEffect()

			v := testcase.Let[int](s, func(t *testcase.T) int {
				t.FailNow() // boom
				return 42
			})

			s.Test(``, func(tcT *testcase.T) {
				tcT.Eventually(func(it *testcase.T) { v.Get(it) })
			})

			stub.Finish()
			s.Finish()

			assert.Must(t).True(stub.IsFailed, `expected to fail`)
		})
		t.Run("temporarily", func(t *testing.T) {
			stub := &doubles.TB{}
			s := testcase.NewSpec(stub)
			s.HasSideEffect()

			failed := testcase.LetValue[bool](s, false)
			counter := testcase.LetValue[int](s, 0)
			expVal := rnd.Int()

			v := testcase.Let[int](s, func(t *testcase.T) int {
				counter.Set(t, counter.Get(t)+1)
				if !failed.Get(t) {
					failed.Set(t, true)
					t.FailNow() // boom
				}
				return expVal
			})

			s.Test(``, func(tcT *testcase.T) {
				tcT.Eventually(func(it *testcase.T) {
					assert.Equal(t, v.Get(it), expVal)
				})
				assert.Equal(t, v.Get(tcT), expVal)
				assert.Equal(t, counter.Get(tcT), 2, "it was expected that the variable init block only run twice, one for failure and one for success")
			})

			stub.Finish()
			s.Finish()

			assert.Must(t).False(stub.IsFailed, `expected to pass`)
		})
	})

	t.Run("a cleanup registered from the eventually block runs with the teardown of the test case", func(t *testing.T) {
		stub := &doubles.TB{}
		s := testcase.NewSpec(stub)
		s.HasSideEffect()

		var cleanups int32
		s.Test(``, func(tcT *testcase.T) {
			tcT.Eventually(func(it *testcase.T) {
				it.Cleanup(func() { atomic.AddInt32(&cleanups, 1) })
			})
			assert.Equal(t, int32(0), atomic.LoadInt32(&cleanups),
				"the cleanup was not expected to run while the test case is still running")
		})

		stub.Finish()
		s.Finish()

		assert.Must(t).True(!stub.IsFailed, `expected to pass`)
		assert.Equal(t, int32(1), atomic.LoadInt32(&cleanups),
			"the cleanup registered from the eventually block was expected to run once the test case is over")
	})

	t.Run("a goroutine started from the eventually block is stopped by the teardown of the test case", func(t *testing.T) {
		stub := &doubles.TB{}
		// stub.Finish cancels the context the goroutine inherits,
		// so it must only happen after the outcome is observed.
		t.Cleanup(stub.Finish)
		s := testcase.NewSpec(stub)
		s.HasSideEffect()

		stopped := make(chan struct{})
		s.Test(``, func(tcT *testcase.T) {
			tcT.Eventually(func(it *testcase.T) {
				it.Go(func(ctx context.Context) error {
					defer close(stopped)
					<-ctx.Done()
					return nil
				})
			})
		})

		s.Finish()

		assert.Must(t).True(!stub.IsFailed, `expected to pass`)
		select {
		case <-stopped:
		default:
			t.Fatal("the goroutine started from the eventually block was expected to be stopped by the teardown of the test case")
		}
	})

	t.Run("the test case can be used from another goroutine while the eventually block is retried", func(t *testing.T) {
		// The point of this case is to give the race detector a chance to catch
		// an eventually block which accesses the test case without synchronisation.
		stub := &doubles.TB{}
		s := testcase.NewSpec(stub)
		s.HasSideEffect()

		const n = 100
		var cleanups int32
		s.Test(``, func(tcT *testcase.T) {
			var registered int32
			tcT.Go(func(ctx context.Context) error {
				for i := 0; i < n; i++ {
					tcT.Cleanup(func() { atomic.AddInt32(&cleanups, 1) })
					atomic.AddInt32(&registered, 1)
					runtime.Gosched()
				}
				return nil
			})
			tcT.Eventually(func(it *testcase.T) {
				assert.Equal(it, int32(n), atomic.LoadInt32(&registered))
			})
		})

		stub.Finish()
		s.Finish()

		assert.Must(t).True(!stub.IsFailed, `expected to pass`)
		assert.Equal(t, int32(n), atomic.LoadInt32(&cleanups))
	})
}

func ExampleNewTWithSpec() {
	s := testcase.NewSpec(nil)
	// some spec specific configuration
	s.Before(func(t *testcase.T) {})

	var tb testing.TB // placeholder
	tc := testcase.NewTWithSpec(tb, s)
	_ = tc
}

func TestNewTWithSpec(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	y := testcase.Var[int]{ID: "Y"}
	v := testcase.Var[int]{
		ID:   "the answer",
		Init: func(t *testcase.T) int { return t.Random.Int() },
	}
	t.Run(`with *Spec`, func(t *testing.T) {
		tb := &doubles.TB{}
		t.Cleanup(tb.Finish)
		s := testcase.NewSpec(tb)
		expectedY := rnd.Int()
		y.LetValue(s, expectedY)
		subject := testcase.NewTWithSpec(tb, s)
		assert.Must(t).Equal(expectedY, y.Get(subject), "use the passed spec's runtime context after set-up")
		assert.Must(t).Equal(v.Get(subject), v.Get(subject), `has test variable cache`)
	})
	t.Run(`without *Spec`, func(t *testing.T) {
		tb := &doubles.TB{}
		t.Cleanup(tb.Finish)
		expectedY := rnd.Int()
		subject := testcase.NewTWithSpec(tb, nil)
		y.Set(subject, expectedY)
		assert.Must(t).Equal(expectedY, y.Get(subject))
		assert.Must(t).Equal(v.Get(subject), v.Get(subject), `has test variable cache`)
	})
	t.Run(`with *testcase.T, same returned`, func(t *testing.T) {
		tb := &doubles.TB{}
		t.Cleanup(tb.Finish)
		tcT1 := testcase.NewTWithSpec(tb, nil)
		tcT2 := testcase.NewTWithSpec(tcT1, nil)
		assert.Must(t).Equal(tcT1, tcT2)
	})
	t.Run(`when nil received, nil is returned`, func(t *testing.T) {
		assert.Must(t).Nil(testcase.NewTWithSpec(nil, nil))
	})
	t.Run(`when NewT is retrieved multiple times, hooks executed only once`, func(t *testing.T) {
		stb := &doubles.TB{}
		s := testcase.NewSpec(stb)
		var out []struct{}
		s.Before(func(t *testcase.T) {
			out = append(out, struct{}{})
		})
		tct := testcase.NewTWithSpec(stb, s)
		tct = testcase.NewTWithSpec(tct, s)
		tct = testcase.NewTWithSpec(tct, s)
		stb.Finish()
		assert.Equal(t, 1, len(out))
	})
}

func ExampleNewT() {
	var tb testing.TB // placeholder
	_ = testcase.NewT(tb)
}

func TestNewT(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	y := testcase.Var[int]{ID: "Y"}
	v := testcase.Var[int]{
		ID:   "the answer",
		Init: func(t *testcase.T) int { return t.Random.Int() },
	}
	t.Run(`smoke`, func(t *testing.T) {
		tb := &doubles.TB{}
		t.Cleanup(tb.Finish)
		expectedY := rnd.Int()
		subject := testcase.NewT(tb)
		y.Set(subject, expectedY)
		assert.Must(t).Equal(expectedY, y.Get(subject))
		assert.Must(t).Equal(v.Get(subject), v.Get(subject), `has test variable cache`)
	})
	t.Run(`with *testcase.T, same returned`, func(t *testing.T) {
		tb := &doubles.TB{}
		t.Cleanup(tb.Finish)
		tcT1 := testcase.NewT(tb)
		tcT2 := testcase.NewT(tcT1)
		assert.Must(t).Equal(tcT1, tcT2)
	})
	t.Run(`when nil received, nil is returned`, func(t *testing.T) {
		assert.Must(t).Nil(testcase.NewT(nil))
	})
}

func BenchmarkT_varDoesNotCountTowardsRun(b *testing.B) {
	s := testcase.NewSpec(b)

	ab := testcase.Let(s, func(t *testcase.T) int {
		time.Sleep(time.Second / 2)
		return t.Random.Int()
	})
	bv := testcase.Let(s, func(t *testcase.T) int {
		_ = ab.Get(t)
		time.Sleep(time.Second / 2)
		return t.Random.Int()
	})
	s.Test(`run`, func(t *testcase.T) {
		// if the benchmark subject is too fast
		// the benchmark goes into a really long measuring loop.
		//
		// expected to perform max around ~ 1001000000 ns/op
		time.Sleep(time.Second)
		_ = bv.Get(t)
	})
}

func TestT_SkipUntil(t *testing.T) {
	const timeLayout = "2006-01-02"
	const skipUntilFormat = "Skip time %s"
	const skipExpiredFormat = "[SkipUntil] expired on %s"
	rnd := random.New(rand.NewSource(time.Now().UnixNano()))
	future := time.Now().AddDate(0, 0, 1)
	t.Run("before SkipUntil deadline, test is skipped", func(t *testing.T) {
		stubTB := &doubles.TB{}
		s := testcase.NewSpec(stubTB)
		var ran bool
		s.Test("", func(t *testcase.T) {
			t.SkipUntil(future.Year(), future.Month(), future.Day(), future.Hour())
			ran = true
		})
		sandbox.Run(func() { s.Finish() })
		assert.Must(t).False(ran)
		assert.Must(t).False(stubTB.LastTB().IsFailed)
		assert.Must(t).True(stubTB.LastTB().IsSkipped)
		assert.Must(t).Contains(stubTB.LastTB().Logs.String(), fmt.Sprintf(skipUntilFormat, future.Format(timeLayout)))
	})
	t.Run("at or after SkipUntil deadline, test is failed", func(t *testing.T) {
		stubTB := &doubles.TB{}
		s := testcase.NewSpec(stubTB)
		today := time.Now().AddDate(0, 0, -1*rnd.IntN(3))
		var ran bool
		s.Test("", func(t *testcase.T) {
			t.SkipUntil(today.Year(), today.Month(), today.Day(), today.Hour())
			ran = true
		})
		sandbox.Run(func() { s.Finish() })
		assert.Must(t).True(ran)
		assert.Must(t).False(stubTB.LastTB().IsFailed)
		assert.Must(t).Contains(stubTB.LastTB().Logs.String(), fmt.Sprintf(skipExpiredFormat, today.Format(timeLayout)))
	})
}

func TestT_UnsetEnv(t *testing.T) {
	const key = "TEST_KEY"
	t.Setenv(key, "this")
	s := testcase.NewSpec(t)
	s.HasSideEffect()
	s.Test("on unset", func(t *testcase.T) {
		t.UnsetEnv(key)
		_, ok := os.LookupEnv(key)
		assert.Must(t).False(ok)
	})
	s.Test("when not used", func(t *testcase.T) {
		_, ok := os.LookupEnv(key)
		assert.Must(t).True(ok)
	})
	s.Finish()

	t.Run("on Parallel test", func(t *testing.T) {
		dtb := &doubles.TB{}

		sandbox.Run(func() {
			defer dtb.Finish()
			defer s.Finish()
			s := testcase.NewSpec(dtb)
			s.Parallel()
			s.Test("on unset it will fail", func(t *testcase.T) {
				t.UnsetEnv(key)
			})
		})

		assert.True(t, dtb.IsFailed)
	})
}

func TestT_SetEnv(t *testing.T) {
	const key = "TEST_KEY"
	defaultValue := "this"
	t.Setenv(key, defaultValue)
	s := testcase.NewSpec(t)
	s.HasSideEffect()
	s.Test("on set", func(t *testcase.T) {
		r := t.Random.StringNC(5, random.CharsetAlpha())
		t.SetEnv(key, r)
		val, ok := os.LookupEnv(key)
		assert.Must(t).True(ok)
		assert.Must(t).Equal(r, val)
	})
	s.Test("on not used", func(t *testcase.T) {
		val, ok := os.LookupEnv(key)
		assert.Must(t).True(ok)
		assert.Must(t).Equal(defaultValue, val)
	})
	s.Finish()
}

func TestT_Setenv(t *testing.T) {
	const key = "TEST_KEY"
	defaultValue := "this"
	t.Setenv(key, defaultValue)
	s := testcase.NewSpec(t)
	s.HasSideEffect()
	s.Test("on set", func(t *testcase.T) {
		r := t.Random.StringNC(5, random.CharsetAlpha())
		t.Setenv(key, r)
		val, ok := os.LookupEnv(key)
		assert.Must(t).True(ok)
		assert.Must(t).Equal(r, val)
	})
	s.Test("on not used", func(t *testcase.T) {
		val, ok := os.LookupEnv(key)
		assert.Must(t).True(ok)
		assert.Must(t).Equal(defaultValue, val)
	})
	s.Finish()
}

func TestT_LogPretty(t *testing.T) {
	dtb := &doubles.TB{}
	tct := testcase.ToT(dtb)
	tct.LogPretty([]int{1, 2, 4})
	type X struct{ Foo string }
	tct.LogPretty(X{Foo: "hello"})
	dtb.Finish()
	assert.Contains(t, dtb.Logs.String(), "[]int{\n\t1,\n\t2,\n\t4,\n}")
	assert.Contains(t, dtb.Logs.String(), "testcase_test.X{\n\tFoo: \"hello\",\n}")
}

func ExampleT_Done() {
	s := testcase.NewSpec(nil)

	s.Test("", func(t *testcase.T) {
		go func() {
			select {
			// case do something for the test
			case <-t.Done():
				return // test is over, time to garbage collect
			}
		}()
	})
}

func TestT_Done(t *testing.T) {
	s := testcase.NewSpec(t)

	var isdone = func(t *testcase.T) bool {
		select {
		case <-t.Done():
			return true
		default:
			return false
		}
	}

	var done int32
	s.Test("", func(t *testcase.T) {
		assert.False(t, isdone(t))
		go func() {
			<-t.Done() // after the test is done
			atomic.AddInt32(&done, 1)
		}()
		t.Cleanup(func() {
			assert.False(t, isdone(t),
				"during cleanup the done should be not ready")

			t.Cleanup(func() {
				assert.False(t, isdone(t),
					"during a cleanup of cleanup, done should not be ready")
			})
		})
	})

	s.Finish()

	assert.Eventually(t, time.Second, func(t testing.TB) {
		assert.Equal(t, atomic.LoadInt32(&done), 1)
	})
}

func TestT_OnFail(t *testing.T) {
	t.Run("on success", func(t *testing.T) {
		dtb := &doubles.TB{}
		var done bool
		s := testcase.NewSpec(dtb)
		s.Test("", func(t *testcase.T) {
			t.OnFail(func() { done = true })
		})
		s.Finish()
		dtb.Finish()
		assert.Equal(t, false, done)
	})
	t.Run("on failure", func(t *testing.T) {
		dtb := &doubles.TB{}
		var done bool
		s := testcase.NewSpec(dtb)
		s.Test("", func(t *testcase.T) {
			t.OnFail(func() { done = true })
			t.FailNow()
		})
		s.Finish()
		dtb.Finish()
		assert.Equal(t, true, done)
	})
	t.Run("race", func(t *testing.T) {
		dtb := &doubles.TB{}
		s := testcase.NewSpec(dtb)
		s.Test("", func(t *testcase.T) {
			testcase.Race(func() {
				t.OnFail(func() {})
			}, func() {
				t.OnFail(func() {})
			})
			t.FailNow()
		})
		s.Finish()
		dtb.Finish()
	})
}

// TestT_Go specifies T#Go, which starts a goroutine that is bound to the
// current test case's execution life-cycle.
//
// The goroutine is expected to run while the test case runs, to be asked to
// stop when the test case ends, and to be waited for by the test case's
// teardown, so that work started by a test never outlives it.
//
// The error returned by the goroutine tells why it stopped. Returning the
// cancellation of its context is how the goroutine reports that it stopped
// because the test ended. That is the framework asking it to stop, not
// something going wrong, so it stays invisible: neither the test case nor the
// test result learns about it.
//
// Any other error is the goroutine telling that something did go wrong.
// The returned job hands it over to the test case, and if the test case doesn't
// join in with Job#Wait to take it over, then it surfaces as a test failure,
// otherwise the test would report success while the work it started failed.
func TestT_Go(t *testing.T) {
	s := testcase.NewSpec(t)

	// outcome describes what happened to a goroutine started with T#Go,
	// observed from outside of the test case which started it.
	type outcome struct {
		// TB is the testing.TB double the test case reported to.
		TB *doubles.TB
		// StartedDuringTest tells if the goroutine began to run
		// while the test case was still running.
		StartedDuringTest bool
		// StoppedDuringTest tells if the goroutine had already stopped
		// by the time the test case's block returned.
		StoppedDuringTest bool
		// RunningAtAfterHook tells if the goroutine was still running
		// while the test case's After hooks were running.
		RunningAtAfterHook bool
		// RunningAfterTeardown tells if the goroutine was still running
		// after the test case's teardown finished.
		RunningAfterTeardown bool
		// TestWasDoneAtStop tells if T#Done was already closed
		// by the time the goroutine stopped.
		TestWasDoneAtStop bool
		// JobDone tells if the Job returned by T#Go reports the goroutine as finished.
		JobDone bool
		// JobErr is what the Job returned by T#Go hands over to the test case.
		//
		// The act collects it only after the test case is over, which is too late
		// to count as the test case joining in on the goroutine.
		JobErr error
	}

	isDone := func(ch <-chan struct{}) bool {
		select {
		case <-ch:
			return true
		default:
			return false
		}
	}

	// shutdownWork is how long a goroutine keeps working after it was asked to
	// stop. It makes the difference observable between a teardown which waits
	// for the goroutine, and one which merely happens to outlive it.
	const shutdownWork = time.Second / 20

	var (
		// goroutine is the block handed over to T#Go.
		// By default, it runs until the test case asks it to stop.
		//
		// It receives the *testcase.T of the test case which started it,
		// which is not the same as the one the assertions run with,
		// hence the tct name, to avoid shadowing the latter by accident.
		goroutine = let.Var(s, func(t *testcase.T) func(*testcase.T, context.Context) error {
			return func(tct *testcase.T, ctx context.Context) error {
				<-ctx.Done()
				time.Sleep(shutdownWork)
				return nil
			}
		})
		// testBody is what the test case does after it started the goroutine.
		// By default, it returns right away and leaves the goroutine to the teardown.
		testBody = let.Var(s, func(t *testcase.T) func(*testcase.T, testcase.Job) {
			return func(tct *testcase.T, job testcase.Job) {}
		})
		// sibling is closed by the additional goroutine which some of the
		// contexts below start next to the one the act observes.
		sibling = let.Var(s, func(t *testcase.T) chan struct{} {
			return make(chan struct{})
		})
	)

	// act runs a test case which starts a goroutine with T#Go,
	// and reports back what happened to that goroutine.
	act := func(t *testcase.T) outcome {
		var (
			out     outcome
			job     testcase.Job
			dtb     = &doubles.TB{}
			blk     = goroutine.Get(t)
			body    = testBody.Get(t)
			started = make(chan struct{})
			stopped = make(chan struct{})
		)
		out.TB = dtb

		spec := testcase.NewSpec(dtb)

		spec.After(func(tct *testcase.T) {
			out.RunningAtAfterHook = !isDone(stopped)
		})

		spec.Test("", func(tct *testcase.T) {
			job = tct.Go(func(ctx context.Context) error {
				close(started)
				defer close(stopped)
				defer func() { out.TestWasDoneAtStop = isDone(tct.Done()) }()
				return blk(tct, ctx)
			})

			select {
			case <-started:
				out.StartedDuringTest = true
			case <-time.After(time.Second):
			}

			// the test case may fail here, so nothing that the act must
			// observe unconditionally is allowed to come after it
			body(tct, job)

			out.StoppedDuringTest = isDone(stopped)
		})

		spec.Finish()
		out.RunningAfterTeardown = !isDone(stopped)
		out.JobDone = isDone(job.Done())
		out.JobErr = job.Wait()
		dtb.Finish()
		return out
	}

	s.Then("the goroutine runs while the test case is running", func(t *testcase.T) {
		assert.True(t, act(t).StartedDuringTest,
			assert.Message("expected that the goroutine already starts while the test case is ongoing"))
	})

	s.Then("the goroutine keeps running until the test case's teardown", func(t *testcase.T) {
		out := act(t)

		assert.False(t, out.StoppedDuringTest,
			assert.Message("expected that the goroutine is not stopped before the test case itself"))
		assert.True(t, out.RunningAtAfterHook,
			assert.Message("expected that the goroutine is not stopped before the test case's After hooks"))
	})

	s.Then("the test case's teardown waits for the goroutine", func(t *testcase.T) {
		assert.False(t, act(t).RunningAfterTeardown,
			assert.Message("expected that without any delay after the test case finished, "+
				"the started goroutine is already stopped, and if needed, waited for"))
	})

	s.Then("T#Done is already closed by the time the goroutine stops", func(t *testcase.T) {
		assert.True(t, act(t).TestWasDoneAtStop,
			assert.Message("expected that T#Done is already closed "+
				"when the goroutine is asked to shut down"))
	})

	s.Then("the returned job reports the goroutine's completion to the test case", func(t *testcase.T) {
		out := act(t)

		assert.True(t, out.JobDone,
			assert.Message("expected that Job#Done is closed once the goroutine stopped"))
		assert.NoError(t, out.JobErr,
			assert.Message("expected that Job#Wait reports what the goroutine returned"))
	})

	s.Then("the test is not failed", func(t *testcase.T) {
		out := act(t)

		assert.False(t, out.TB.IsFailed,
			assert.Message("expected that the goroutine's life-cycle is not registered as an issue"),
			assert.MessageF("%s", out.TB.Logs.String()))
	})

	s.When("the test case fails before it returns", func(s *testcase.Spec) {
		testBody.Let(s, func(t *testcase.T) func(*testcase.T, testcase.Job) {
			return func(tct *testcase.T, job testcase.Job) {
				tct.FailNow()
			}
		})

		s.Then("the goroutine is still stopped and waited for", func(t *testcase.T) {
			assert.False(t, act(t).RunningAfterTeardown,
				assert.Message("expected that a failing test case doesn't leak the goroutine it started"))
		})
	})

	s.When("the test case starts more than one goroutine", func(s *testcase.Spec) {
		testBody.Let(s, func(t *testcase.T) func(*testcase.T, testcase.Job) {
			stopped := sibling.Get(t)
			return func(tct *testcase.T, job testcase.Job) {
				tct.Go(func(ctx context.Context) error {
					defer close(stopped)
					<-ctx.Done()
					time.Sleep(shutdownWork)
					return nil
				})
			}
		})

		s.Then("the test case's teardown waits for all of them", func(t *testcase.T) {
			assert.False(t, act(t).RunningAfterTeardown,
				assert.Message("expected that the first goroutine is waited for"))
			assert.True(t, isDone(sibling.Get(t)),
				assert.Message("expected that every goroutine is waited for, not only the first one"))
		})
	})

	s.When("the test case stops the goroutine early with the returned job", func(s *testcase.Spec) {
		testBody.Let(s, func(t *testcase.T) func(*testcase.T, testcase.Job) {
			return func(tct *testcase.T, job testcase.Job) {
				job.Cancel()
				_ = job.Wait()
			}
		})

		goroutine.Let(s, func(t *testcase.T) func(*testcase.T, context.Context) error {
			return func(tct *testcase.T, ctx context.Context) error {
				<-ctx.Done()
				// keep working for a bit, so a Job#Wait which fails to wait
				// is distinguishable from one that waits properly
				time.Sleep(shutdownWork)
				return ctx.Err()
			}
		})

		s.Then("the goroutine is already stopped when the test case returns", func(t *testcase.T) {
			assert.True(t, act(t).StoppedDuringTest,
				assert.Message("expected that the goroutine is stopped due to the use of Cancel+Wait"))
		})

		s.Then("the test is not failed", func(t *testcase.T) {
			out := act(t)

			assert.False(t, out.TB.IsFailed,
				assert.Message("expected that an early termination requested by the test case is not an issue"),
				assert.MessageF("%s", out.TB.Logs.String()))
		})

		s.Then("the returned job doesn't hand over the cancellation as an error", func(t *testcase.T) {
			assert.NoError(t, act(t).JobErr,
				assert.Message("expected that the test case's own cancellation "+
					"is not reported back to it as if something went wrong"))
		})
	})

	s.When("the goroutine fails with T#Fatal", func(s *testcase.Spec) {
		goroutine.Let(s, func(t *testcase.T) func(*testcase.T, context.Context) error {
			return func(tct *testcase.T, ctx context.Context) error {
				tct.Fatal("boom")
				return nil
			}
		})

		s.Then("the test is failed", func(t *testcase.T) {
			assert.True(t, act(t).TB.IsFailed,
				assert.Message("expected that a failure raised in the goroutine fails the test case"))
		})
	})

	s.When("the goroutine ends abruptly, without returning", func(s *testcase.Spec) {
		goroutine.Let(s, func(t *testcase.T) func(*testcase.T, context.Context) error {
			return func(tct *testcase.T, ctx context.Context) error {
				runtime.Goexit()
				return nil
			}
		})

		// The goroutine never returns, so it has no error to hand over,
		// and the test case has nothing to take over with Job#Wait either.
		s.Then("the test is failed", func(t *testcase.T) {
			assert.True(t, act(t).TB.IsFailed,
				assert.Message("expected that a goroutine which vanished without a word "+
					"is not silently accepted"))
		})
	})

	s.When("the goroutine stops with an error", func(s *testcase.Spec) {
		expectedErr := let.Error(s).
			EagerLoading(s)

		goroutine.Let(s, func(t *testcase.T) func(*testcase.T, context.Context) error {
			err := expectedErr.Get(t)
			return func(tct *testcase.T, ctx context.Context) error {
				<-ctx.Done()
				return err
			}
		})

		s.Then("the test is failed", func(t *testcase.T) {
			assert.True(t, act(t).TB.IsFailed,
				assert.Message("expected that an error returned by the goroutine is reported as a test failure"))
		})

		s.Then("the error is reported", func(t *testcase.T) {
			assert.Contains(t, act(t).TB.Logs.String(), expectedErr.Get(t).Error())
		})

		s.Then("the returned job hands the error over to the test case", func(t *testcase.T) {
			assert.ErrorIs(t, act(t).JobErr, expectedErr.Get(t))
		})

		// Without joining in, the goroutine is only scheduled during the
		// teardown, by which time its context is already cancelled. That
		// difference is what decides whether a returned error can be mistaken
		// for the cancellation.
		s.And("it stops while the test case is still running", func(s *testcase.Spec) {
			goroutine.Let(s, func(t *testcase.T) func(*testcase.T, context.Context) error {
				err := expectedErr.Get(t)
				return func(tct *testcase.T, ctx context.Context) error {
					return err
				}
			})

			// Job#Done only observes that the goroutine is over,
			// it doesn't take over the error like Job#Wait would.
			testBody.Let(s, func(t *testcase.T) func(*testcase.T, testcase.Job) {
				return func(tct *testcase.T, job testcase.Job) {
					<-job.Done()
				}
			})

			s.Then("the test is failed", func(t *testcase.T) {
				assert.True(t, act(t).TB.IsFailed,
					assert.Message("expected that an error is reported as a test failure "+
						"regardless of whether the test's context is already cancelled"))
			})
		})

		s.And("the test case joins in on the goroutine with Job#Wait", func(s *testcase.Spec) {
			goroutine.Let(s, func(t *testcase.T) func(*testcase.T, context.Context) error {
				err := expectedErr.Get(t)
				return func(tct *testcase.T, ctx context.Context) error {
					return err
				}
			})

			testBody.Let(s, func(t *testcase.T) func(*testcase.T, testcase.Job) {
				return func(tct *testcase.T, job testcase.Job) {
					_ = job.Wait()
				}
			})

			s.Then("the test is not failed", func(t *testcase.T) {
				out := act(t)

				assert.False(t, out.TB.IsFailed,
					assert.Message("expected that by joining in, the test case takes over the error, "+
						"and it is up to the test case to decide what it means"),
					assert.MessageF("%s", out.TB.Logs.String()))
			})

			s.Then("the error is not reported", func(t *testcase.T) {
				assert.NotContains(t, act(t).TB.Logs.String(), expectedErr.Get(t).Error())
			})

			s.Then("the returned job hands the error over to the test case", func(t *testcase.T) {
				assert.ErrorIs(t, act(t).JobErr, expectedErr.Get(t))
			})
		})

		// The verdict can only be final once every cleanup had its chance to
		// join in, which is why it is passed at the very end of the teardown.
		s.And("the test case joins in on the goroutine from a Cleanup", func(s *testcase.Spec) {
			goroutine.Let(s, func(t *testcase.T) func(*testcase.T, context.Context) error {
				err := expectedErr.Get(t)
				return func(tct *testcase.T, ctx context.Context) error {
					return err
				}
			})

			testBody.Let(s, func(t *testcase.T) func(*testcase.T, testcase.Job) {
				return func(tct *testcase.T, job testcase.Job) {
					tct.Cleanup(func() { _ = job.Wait() })
				}
			})

			s.Then("the test is not failed", func(t *testcase.T) {
				out := act(t)

				assert.False(t, out.TB.IsFailed,
					assert.Message("expected that joining in from a cleanup takes over the error just the same"),
					assert.MessageF("%s", out.TB.Logs.String()))
			})
		})

		// A cleanup may be registered by a goroutine which is already shutting
		// down, so it only reaches the teardown queue while the goroutines are
		// winding down. Such a cleanup is still the test case's own chance to
		// join in, so the verdict may only be passed once the teardown queue is
		// drained, and not merely once the goroutines are over.
		s.And("a Cleanup registered during the shutdown joins in on the goroutine", func(s *testcase.Spec) {
			goroutine.Let(s, func(t *testcase.T) func(*testcase.T, context.Context) error {
				err := expectedErr.Get(t)
				return func(tct *testcase.T, ctx context.Context) error {
					// stops right away, so by the time the teardown begins,
					// its error is already there, waiting to be collected
					return err
				}
			})

			testBody.Let(s, func(t *testcase.T) func(*testcase.T, testcase.Job) {
				return func(tct *testcase.T, job testcase.Job) {
					tct.Go(func(ctx context.Context) error {
						<-ctx.Done()
						tct.Cleanup(func() {
							// teardown work which has to be done before joining in,
							// so the goroutines are surely all over by then
							time.Sleep(shutdownWork)
							_ = job.Wait()
						})
						return nil
					})
				}
			})

			s.Then("the test is not failed", func(t *testcase.T) {
				out := act(t)

				assert.False(t, out.TB.IsFailed,
					assert.Message("expected that a cleanup which is registered while the goroutines "+
						"are already shutting down still takes the error over in time"),
					assert.MessageF("%s", out.TB.Logs.String()))
			})

			s.Then("the error is not reported", func(t *testcase.T) {
				assert.NotContains(t, act(t).TB.Logs.String(), expectedErr.Get(t).Error())
			})
		})
	})

	s.When("more than one goroutine stops with an error", func(s *testcase.Spec) {
		var (
			firstErr = let.Var(s, func(t *testcase.T) error {
				return fmt.Errorf("the first goroutine failed: %s", t.Random.UUID())
			}).EagerLoading(s)
			secondErr = let.Var(s, func(t *testcase.T) error {
				return fmt.Errorf("the second goroutine failed: %s", t.Random.UUID())
			}).EagerLoading(s)
		)

		goroutine.Let(s, func(t *testcase.T) func(*testcase.T, context.Context) error {
			err := firstErr.Get(t)
			return func(tct *testcase.T, ctx context.Context) error {
				return err
			}
		})

		testBody.Let(s, func(t *testcase.T) func(*testcase.T, testcase.Job) {
			err := secondErr.Get(t)
			return func(tct *testcase.T, job testcase.Job) {
				tct.Go(func(ctx context.Context) error { return err })
			}
		})

		s.Then("every error is reported", func(t *testcase.T) {
			logs := act(t).TB.Logs.String()

			assert.Contains(t, logs, firstErr.Get(t).Error())
			assert.Contains(t, logs, secondErr.Get(t).Error(),
				assert.Message("expected that a goroutine's error is not shadowed by another one's"))
		})
	})

	s.When("the goroutine stops because the test case ended", func(s *testcase.Spec) {
		goroutine.Let(s, func(t *testcase.T) func(*testcase.T, context.Context) error {
			return func(tct *testcase.T, ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			}
		})

		s.Then("the test is not failed", func(t *testcase.T) {
			out := act(t)

			assert.False(t, out.TB.IsFailed,
				assert.Message("expected that the cancellation which ends the goroutine is not an issue"),
				assert.MessageF("%s", out.TB.Logs.String()))
		})

		s.Then("the returned job doesn't hand over the cancellation as an error", func(t *testcase.T) {
			assert.NoError(t, act(t).JobErr,
				assert.Message("expected that the framework's own cancellation "+
					"is not reported back to the test case as if something went wrong"))
		})

		s.And("the cancellation is wrapped into a more descriptive error", func(s *testcase.Spec) {
			goroutine.Let(s, func(t *testcase.T) func(*testcase.T, context.Context) error {
				return func(tct *testcase.T, ctx context.Context) error {
					<-ctx.Done()
					return fmt.Errorf("worker stopped: %w", ctx.Err())
				}
			})

			s.Then("the test is not failed", func(t *testcase.T) {
				out := act(t)

				assert.False(t, out.TB.IsFailed,
					assert.Message("expected that the cancellation is recognised through the error wrapping"),
					assert.MessageF("%s", out.TB.Logs.String()))
			})

			s.Then("the returned job doesn't hand over the wrapped cancellation as an error", func(t *testcase.T) {
				assert.NoError(t, act(t).JobErr,
					assert.Message("expected that the cancellation is recognised through the error wrapping"))
			})
		})
	})
}

// TestT_finish_lateCleanupRace reproduces a hang in the test framework's
// test teardown. A goroutine spawned via t.Go from a Before hook registers
// a t.Cleanup callback *after* the test body has returned. The current
// finish() drains the cleanup queue once, *before* waiting on the
// goroutines to exit, so the late-registered cleanup is silently lost.
//
// If the goroutine is blocked on something that the cleanup would unblock
// (a channel close, a phaser finish, a condition variable broadcast, etc.)
// the goroutine never exits, t.g.Wait() blocks, and the test only ends
// when the Go test timeout panic fires.
//
// Why this matters: the framework encourages parallel goroutines via
// t.Go. Any pattern where a goroutine does its own t.Cleanup
// registration (a common shape — e.g. allocating a resource lazily
// inside the goroutine and registering a finalizer for it) silently
// hangs the test, even though the same code in a single-goroutine
// pattern works fine. The fix should let late cleanups reach the
// teardown queue while t.g.Wait is still active.
func TestT_Go_finish_lateCleanupRace(t *testing.T) {
	const failFastBudget = 5 * time.Second

	s := testcase.NewSpec(t)

	// Released on test teardown. The test goroutine never touches this
	// channel directly; only the worker goroutine waits on it, and only
	// the t.Cleanup callback closes it.
	release := make(chan struct{})

	// Tracks whether the worker goroutine observed the channel close.
	// Used to prove the cleanup ran (and not just that the test exited
	// because of the timeout alarm).
	var workerObservedClose atomic.Bool

	s.Before(func(t *testcase.T) {
		t.Go(func(ctx context.Context) error {
			// Simulate a resource allocated lazily inside the goroutine
			// that needs teardown. The t.Cleanup call below runs in
			// the worker goroutine, after the test body has returned.
			t.Cleanup(func() {
				// If finish() drained the queue before we registered,
				// this close is never observed by the worker and the
				// test hangs on t.g.Wait.
				close(release)
			})
			<-release
			workerObservedClose.Store(true)
			return nil
		})
	})

	s.Test("a goroutine that registers t.Cleanup after the test body returns must not hang on teardown", func(t *testcase.T) {
		// Body returns immediately. The worker goroutine is parked on
		// <-release. The race window is the gap between this return and
		// t.g.Wait(): if t.Cleanup hasn't been registered yet, the close
		// is never invoked and the worker hangs forever.
	})

	// The whole point of this assertion: finish the spec within the
	// fail-fast budget. If the bug is present, this blocks until the
	// Go test timeout fires, which prints a goroutine dump and fails
	// the test much later — and from a flaky signal that the operator
	// would have to chase down. With the fix, finish() returns as soon
	// as the worker observes the close.
	assert.Within(t, failFastBudget, func(ctx context.Context) {
		s.Finish()
	})

	// Sanity check: the worker actually observed the close. If
	// t.Cleanup was lost, this is false; if it ran, this is true.
	assert.True(t, workerObservedClose.Load(),
		assert.Message("expected the worker's t.Cleanup to run during teardown; "+
			"if this is false, the framework silently dropped a late-registered cleanup "+
			"and the test only finished via the Go test-timeout alarm"))
}

// TestT_Go_finish_lateCleanupIsNotDropped is the non-blocking half of
// TestT_Go_finish_lateCleanupRace.
//
// The goroutine registers its cleanup strictly after the test body returned,
// but it does not depend on that cleanup to be able to exit. Nothing hangs
// here, so the defect it covers is not the deadlock but the silent data loss
// that causes it: a cleanup that reaches the queue after the drain is simply
// never executed, and whatever resource it was meant to release leaks.
//
// In contrast to the deadlock, this is observable without any timing budget,
// which makes it the cheaper and more precise regression guard of the two.
func TestT_Go_finish_lateCleanupIsNotDropped(t *testing.T) {
	s := testcase.NewSpec(t)

	var cleanupRan int32

	s.Before(func(t *testcase.T) {
		t.Go(func(ctx context.Context) error {
			// The context is cancelled by the teardown, after it already
			// drained the queue once, so the registration below is
			// guaranteed to be a late one.
			<-ctx.Done()
			t.Cleanup(func() { atomic.StoreInt32(&cleanupRan, 1) })
			return nil
		})
	})

	s.Test("a cleanup registered by a goroutine after the test body returned still runs", func(t *testcase.T) {})

	s.Finish()

	assert.Equal(t, int32(1), atomic.LoadInt32(&cleanupRan),
		assert.Message("expected that a cleanup registered while the goroutines are "+
			"shutting down is still executed, instead of being silently dropped"))
}

// TestT_Go_finish_addsNoFixedLatencyPerTestCase guards the *way* the late
// cleanups are picked up.
//
// Keeping the teardown queue drained while the goroutines wind down can be
// implemented either by waiting for a registration to be signalled, or by
// waking up periodically to look for one. Both make the tests above pass, but
// a periodic wake-up adds its full interval to the teardown of every single
// test case, because the wake-up happens before the teardown can observe that
// the goroutines are already gone.
//
// That cost is easy to miss: Spec#Benchmark runs the teardown once per
// iteration with the timer stopped, so it never shows up in the reported
// ns/op, it only inflates the wall clock, and `go test` keeps raising b.N
// based on the ns/op it measured.
//
// A single test case cannot tell a signal from a poll reliably, so the budget
// below is spread over many of them.
func TestT_Go_finish_addsNoFixedLatencyPerTestCase(t *testing.T) {
	const (
		testCases = 500
		// Even a 1ms wake-up interval, the smallest that is still practical,
		// needs ~1.3ms per test case, so >= 650ms for the test cases below,
		// while reacting to a signal keeps them in the 15-30ms range.
		budget = 200 * time.Millisecond
	)

	var dtb doubles.TB

	// A Spec made by NewSpec is sync, so its test cases already run as they
	// are declared, and the measured window has to cover that too.
	startedAt := time.Now()

	s := testcase.NewSpec(&dtb)
	for i := 0; i < testCases; i++ {
		s.Test("", func(t *testcase.T) {
			// the teardown has to wait for this goroutine,
			// while staying ready for a cleanup it may still register
			t.Go(func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			})
		})
	}
	s.Finish()

	duration := time.Since(startedAt)

	assert.False(t, dtb.IsFailed, assert.MessageF("%s", dtb.Logs.String()))

	assert.True(t, duration < budget,
		assert.MessageF("%d test cases took %s, which is over the %s budget (%s per test case).\n"+
			"The teardown is expected to react to the goroutines finishing, not to poll for it, "+
			"because a poll interval is paid by every single test case.",
			testCases, duration, budget, duration/testCases))
}

// TestT_Go_finish_panicIsRaisedOnTheTestsGoroutine pins down where a panic
// that happened inside a T#Go goroutine is re-raised.
//
// tcsync.Group captures such a panic and replays it on whoever calls
// Group#Wait. That call has to stay on the test's own goroutine: only there
// can the testing framework recover it, mark the test as failed, print the
// TESTCASE_SEED and name the test that caused it. Moved onto a helper
// goroutine, the very same panic becomes unrecoverable and takes down the
// whole test binary with a stack trace that points at the framework's
// internals instead of the offending test.
//
// The sandbox below can only observe the panic if it is raised on the
// goroutine the sandbox owns. Should the teardown ever start raising it
// elsewhere, there is no recover that could catch it, and this test fails the
// only way it still can: by crashing the test binary.
func TestT_Go_finish_panicIsRaisedOnTheTestsGoroutine(t *testing.T) {
	const panicValue = "boom from a T#Go goroutine"

	var dtb doubles.TB
	tct := testcase.NewT(&dtb)
	tct.Go(func(ctx context.Context) error { panic(panicValue) })

	// doubles.TB#Finish runs the teardown that testcase.NewT registered,
	// on this very goroutine, the same way *testing.T would run it.
	out := sandbox.Run(func() { dtb.Finish() })

	assert.True(t, out.Panic,
		assert.Message("expected that the panic of a T#Go goroutine is replayed "+
			"on the goroutine that runs the test's teardown"))

	assert.Contains(t, fmt.Sprint(out.PanicValue), panicValue,
		assert.Message("expected that the original panic value is preserved"))
}
