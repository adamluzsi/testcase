package testcase_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/contracts"
	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/sandbox"

	"github.com/adamluzsi/testcase/random"

	"github.com/adamluzsi/testcase"
)

var _ testing.TB = &testcase.T{}

func TestT_implementsTestingTB(t *testing.T) {
	testcase.RunSuite(t, contracts.TestingTB{
		Subject: func(t *testcase.T) testing.TB {
			stub := &doubles.TB{}
			t.Cleanup(stub.Finish)
			return testcase.NewT(stub, nil)
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
				t.Must.Equal(n.Get(t), nog.Get(t))
				t.Must.Equal(m.Get(t), mog.Get(t))
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
			t.Must.Equal(initValue+2, x.Get(t))
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
						t.Must.Equal(42, wDefer.Get(t))
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
//goland:noinspection GoDeferGo
func TestT_Defer_failNowWillNotHang(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer recover()
		s := testcase.NewSpec(&doubles.RecorderTB{})

		s.Before(func(t *testcase.T) {
			t.Defer(func() { t.FailNow() })
		})

		s.Context(``, func(s *testcase.Spec) {
			s.Test(``, func(t *testcase.T) {
				panic(`die`)
			})
		})

		s.Test(``, func(t *testcase.T) {})
	}()
	wg.Wait()
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
		t.Must.Panic(func() { _ = v.Get(t) })
	})

	s.Test(`panic message`, func(t *testcase.T) {
		message := getPanicMessage(func() { _ = v.Get(t) })
		t.Must.Contain(message, `/testcase/T_test.go`)
		t.Must.Contain(message, `expected 1`)
		t.Must.Contain(message, `got 2`)
	})

	s.Test(`interface type with wrong implementation`, func(t *testcase.T) {
		type notContextForSure struct{}
		var fn = func(ctx context.Context) {}
		t.Must.Panic(func() { t.Defer(fn, notContextForSure{}) })
		message := getPanicMessage(func() { t.Defer(fn, notContextForSure{}) })
		t.Must.Contain(message, `/testcase/T_test.go`)
		t.Must.Contain(message, `doesn't implements context.Context`)
		t.Must.Contain(message, `argument[0]`)
	})
}

func TestT_Defer_withArgumentsButArgumentTypeMismatch(t *testing.T) {
	s := testcase.NewSpec(t)

	v := testcase.Let(s, func(t *testcase.T) int {
		t.Defer(func(n int) {}, `this is not ok`)
		return 42
	})

	s.Test(`testCase that it will panics early on to help ease the pain of seeing mistakes`, func(t *testcase.T) {
		t.Must.Panic(func() { _ = v.Get(t) })
	})

	s.Test(`panic message`, func(t *testcase.T) {
		message := func() (r string) {
			defer func() { r = recover().(string) }()
			_ = v.Get(t)
			return ``
		}()

		t.Must.Contain(message, `/testcase/T_test.go`)
		t.Must.Contain(message, `expected int`)
		t.Must.Contain(message, `got string`)
	})
}

func TestT_TB(t *testing.T) {
	s := testcase.NewSpec(t)

	for i := 0; i < 10; i++ {
		var ts []testing.TB
		s.Test(`*testcase.TB is set to the given testcase's *testing.T`, func(t *testcase.T) {
			t.Must.NotNil(t.TB)
			t.Must.NotContain(ts, t.TB, `TB should be unique for each testCase run`)
			ts = append(ts, t.TB)
		})
	}
}

func TestT_Defer_calledWithoutFunctionAndWillPanic(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Test(`defer expected to panic for non function input as first value`, func(t *testcase.T) {
		var withReturnValue = func() int { return 42 }

		t.Must.Panic(func() { t.Defer(withReturnValue()) })
	})

	s.Test(`defer expected to panic for invalid inputs`, func(t *testcase.T) {
		var dummyClose = func() error { return nil }
		pv := t.Must.Panic(func() { t.Defer(dummyClose()) })
		t.Must.Contain(pv, `T#Defer can only take functions`)
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
					t.Must.True(t.HasTag(`a`))
					t.Must.True(t.HasTag(`b`))
					t.Must.True(t.HasTag(`c`))
					t.Must.True(!t.HasTag(`d`))
				})
			})

			s.Test(`b`, func(t *testcase.T) {
				t.Must.True(t.HasTag(`a`))
				t.Must.True(t.HasTag(`b`))
				t.Must.True(!t.HasTag(`c`))
				t.Must.True(!t.HasTag(`d`))
			})
		})

		s.Test(`a`, func(t *testcase.T) {
			t.Must.True(t.HasTag(`a`))
			t.Must.True(!t.HasTag(`b`))
			t.Must.True(!t.HasTag(`c`))
			t.Must.True(!t.HasTag(`d`))
		})
	})

	s.Test(``, func(t *testcase.T) {
		t.Must.True(!t.HasTag(`a`))
		t.Must.True(!t.HasTag(`b`))
		t.Must.True(!t.HasTag(`c`))
		t.Must.True(!t.HasTag(`d`))
	})
}

func TestT_Random(t *testing.T) {
	randomGenerationWorks := func(t *testcase.T) {
		assert.Eventually{RetryStrategy: assert.Waiter{WaitDuration: time.Second}}.Assert(t, func(it assert.It) {
			it.Must.True(0 < t.Random.Int())
		})
	}

	t.Run(`when environment value is set`, func(t *testing.T) {
		testcase.SetEnv(t, testcase.EnvKeySeed, `42`)
		s := testcase.NewSpec(t)
		s.Test(``, func(t *testcase.T) {
			t.Must.Equal(random.New(rand.NewSource(42)), t.Random)

			randomGenerationWorks(t)
		})
	})

	s := testcase.NewSpec(t)
	s.Test(``, func(t *testcase.T) {
		randomGenerationWorks(t)
	})
}

func TestT_Eventually(t *testing.T) {
	t.Run(`with default eventually retry strategy`, func(t *testing.T) {
		stub := &doubles.TB{}
		s := testcase.NewSpec(stub)
		s.HasSideEffect()
		var eventuallyRan bool
		s.Test(``, func(t *testcase.T) {
			t.Eventually(func(it assert.It) {
				eventuallyRan = true
				it.Must.True(t.Random.Bool())
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
		strategy := assert.RetryStrategyFunc(func(condition func() bool) {
			strategyUsed = true
			for condition() {
			}
		})
		s := testcase.NewSpec(stub, testcase.RetryStrategyForEventually(strategy))
		s.HasSideEffect()
		s.Test(``, func(t *testcase.T) {
			t.Eventually(func(it assert.It) {
				it.Must.True(t.Random.Bool())
			}) // eventually pass
		})
		stub.Finish()
		s.Finish()
		assert.Must(t).True(!stub.IsFailed, `expected to pass`)
		assert.Must(t).True(strategyUsed, `retry strategy of the eventually call was used`)
	})
}

func TestNewT(t *testing.T) {
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
		subject := testcase.NewT(tb, s)
		assert.Must(t).Equal(expectedY, y.Get(subject), "use the passed spec's runtime context after set-up")
		assert.Must(t).Equal(v.Get(subject), v.Get(subject), `has test variable cache`)
	})
	t.Run(`without *Spec`, func(t *testing.T) {
		tb := &doubles.TB{}
		t.Cleanup(tb.Finish)
		expectedY := rnd.Int()
		subject := testcase.NewT(tb, nil)
		y.Set(subject, expectedY)
		assert.Must(t).Equal(expectedY, y.Get(subject))
		assert.Must(t).Equal(v.Get(subject), v.Get(subject), `has test variable cache`)
	})
	t.Run(`with *testcase.T, same returned`, func(t *testing.T) {
		tb := &doubles.TB{}
		t.Cleanup(tb.Finish)
		tcT1 := testcase.NewT(tb, nil)
		tcT2 := testcase.NewT(tcT1, nil)
		assert.Must(t).Equal(tcT1, tcT2)
	})
	t.Run(`when nil received, nil is returned`, func(t *testing.T) {
		assert.Must(t).Nil(testcase.NewT(nil, nil))
	})
	t.Run(`when NewT is retrieved multiple times, hooks executed only once`, func(t *testing.T) {
		stb := &doubles.TB{}
		s := testcase.NewSpec(stb)
		var out []struct{}
		s.Before(func(t *testcase.T) {
			out = append(out, struct{}{})
		})
		tct := testcase.NewT(stb, s)
		tct = testcase.NewT(tct, s)
		tct = testcase.NewT(tct, s)
		stb.Finish()
		assert.Equal(t, 1, len(out))
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
	const skipExpiredFormat = "Skip expired on %s"
	rnd := random.New(rand.NewSource(time.Now().UnixNano()))
	future := time.Now().AddDate(0, 0, 1)
	t.Run("before SkipUntil deadline, test is skipped", func(t *testing.T) {
		stubTB := &doubles.TB{}
		s := testcase.NewSpec(stubTB)
		var ran bool
		s.Test("", func(t *testcase.T) {
			t.SkipUntil(future.Year(), future.Month(), future.Day())
			ran = true
		})
		sandbox.Run(func() { s.Finish() })
		assert.Must(t).False(ran)
		assert.Must(t).False(stubTB.IsFailed)
		assert.Must(t).True(stubTB.IsSkipped)
		assert.Must(t).Contain(stubTB.Logs.String(), fmt.Sprintf(skipUntilFormat, future.Format(timeLayout)))
	})
	t.Run("at or after SkipUntil deadline, test is failed", func(t *testing.T) {
		stubTB := &doubles.TB{}
		s := testcase.NewSpec(stubTB)
		today := time.Now().AddDate(0, 0, -1*rnd.IntN(3))
		var ran bool
		s.Test("", func(t *testcase.T) {
			t.SkipUntil(today.Year(), today.Month(), today.Day())
			ran = true
		})
		sandbox.Run(func() { s.Finish() })
		assert.Must(t).False(ran)
		assert.Must(t).True(stubTB.IsFailed)
		assert.Must(t).Contain(stubTB.Logs.String(), fmt.Sprintf(skipExpiredFormat, today.Format(timeLayout)))
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
		t.Must.False(ok)
	})
	s.Test("when not used", func(t *testcase.T) {
		_, ok := os.LookupEnv(key)
		t.Must.True(ok)
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
		t.Must.True(ok)
		t.Must.Equal(r, val)
	})
	s.Test("on not used", func(t *testcase.T) {
		val, ok := os.LookupEnv(key)
		t.Must.True(ok)
		t.Must.Equal(defaultValue, val)
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
		t.Must.True(ok)
		t.Must.Equal(r, val)
	})
	s.Test("on not used", func(t *testcase.T) {
		val, ok := os.LookupEnv(key)
		t.Must.True(ok)
		t.Must.Equal(defaultValue, val)
	})
	s.Finish()
}
