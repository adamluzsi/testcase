package let_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/random"
	"go.llib.dev/testcase/random/sextype"
	"go.llib.dev/testcase/sandbox"
)

func TestWith(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	t.Run("func() V", func(t *testing.T) {
		s := testcase.NewSpec(t)
		n := rnd.Int()
		v := let.With[int](s, func() int {
			return n
		})
		s.Test("", func(t *testcase.T) {
			t.Must.Equal(n, v.Get(t))
		})
	})
	t.Run("func(testing.TB) V", func(t *testing.T) {
		s := testcase.NewSpec(t)
		n := rnd.String()
		v := let.With[string](s, func(testing.TB) string {
			return n
		})
		s.Test("", func(t *testcase.T) {
			t.Must.Equal(n, v.Get(t))
		})
	})
	t.Run("func(*testcase.T) V", func(t *testing.T) {
		s := testcase.NewSpec(t)
		n := let.UUID(s)
		v := let.With[string](s, func(t *testcase.T) string {
			return n.Get(t)
		})
		s.Test("", func(t *testcase.T) {
			t.Must.Equal(n.Get(t), v.Get(t))
		})
	})
}

func TestAs(t *testing.T) {
	t.Run("primitive type", func(t *testing.T) {
		type MyString string

		s := testcase.NewSpec(t)
		v1 := let.String(s)
		v2 := let.As[MyString](v1)

		s.Test("", func(t *testcase.T) {
			t.Must.Equal(MyString(v1.Get(t)), v2.Get(t))
		})
	})

	t.Run("interface type", func(t *testing.T) {
		type TimeAfterer interface {
			After(u time.Time) bool
		}

		s := testcase.NewSpec(t)
		v1 := let.Time(s)
		v2 := let.As[TimeAfterer](v1)

		s.Test("", func(t *testcase.T) {
			t.Must.Equal(TimeAfterer(v1.Get(t)), v2.Get(t))
		})
	})

	t.Run("panics on incorrect conversation", func(t *testing.T) {
		ro := sandbox.Run(func() {
			s := testcase.NewSpec(t)
			v1 := let.Time(s)
			_ = let.As[string](v1)
		})
		assert.False(t, ro.OK)
		assert.False(t, ro.Goexit)
		assert.NotNil(t, ro.PanicValue)
	})
}

var rnd = random.New(random.CryptoSeed{})

func Test_smoke(t *testing.T) {
	s := testcase.NewSpec(t)

	Context := let.Context(s)
	Error := let.Error(s)
	String := let.String(s)
	StringNC := let.StringNC(s, 42, random.CharsetASCII())
	Bool := let.Bool(s)
	Int := let.Int(s)
	IntN := let.IntN(s, 42)
	IntNalt := let.IntN(s, 1000)
	IntB := let.IntB(s, 7, 42)
	Time := let.Time(s)
	TimeB := let.TimeB(s, time.Now().AddDate(-1, 0, 0), time.Now())

	lenHexN := rnd.IntBetween(1, 7)
	HexN := let.HexN(s, lenHexN)
	UUID := let.UUID(s)
	Element := let.OneOf(s, "foo", "bar", "baz")
	DurationBetween := let.DurationBetween(s, time.Second, time.Minute)
	recorder := let.HTTPTestResponseRecorder(s)

	charsterIs := func(t *testcase.T, cs, str string) {
		for _, v := range str {
			t.Must.Contains(cs, string(v))
		}
	}

	s.Test("", func(t *testcase.T) {
		t.Must.NotNil(Context.Get(t))
		t.Must.NoError(Context.Get(t).Err())
		t.Must.NotWithin(time.Millisecond, func(ctx context.Context) {
			select {
			case <-Context.Get(t).Done():
				// expect to block
			case <-ctx.Done():
				// will be done after the assertion
			}
		})
		t.Must.Error(Error.Get(t))
		t.Must.NotEmpty(String.Get(t))
		t.Must.NotEmpty(StringNC.Get(t))
		t.Must.True(len(StringNC.Get(t)) == 42)
		charsterIs(t, random.CharsetASCII(), StringNC.Get(t))
		t.Must.NotEmpty(Int.Get(t))
		assert.AnyOf(t, func(a *assert.A) {
			a.Test(func(it testing.TB) { assert.NotEmpty(it, IntN.Get(t)) })
			a.Test(func(it testing.TB) { assert.NotEmpty(it, IntNalt.Get(t)) })
		})
		t.Must.NotEmpty(IntB.Get(t))
		t.Must.NotEmpty(DurationBetween.Get(t))
		t.Must.True(time.Second <= DurationBetween.Get(t))
		t.Must.True(DurationBetween.Get(t) <= time.Minute)
		t.Must.NotEmpty(Time.Get(t))
		t.Must.NotEmpty(TimeB.Get(t))
		t.Must.True(TimeB.Get(t).After(time.Now().AddDate(-1, 0, -1)))
		t.Must.NotEmpty(UUID.Get(t))
		t.Must.NotEmpty(HexN.Get(t))
		t.Must.Equal(len(HexN.Get(t)), lenHexN)
		t.Must.NotEmpty(Element.Get(t))
		t.Eventually(func(it *testcase.T) {
			it.Must.True(Bool.Get(testcase.ToT(&t.TB)))
		})
		assert.NotNil(t, recorder.Get(t))
		recorder.Get(t).WriteHeader(http.StatusTeapot)
	})
}

func TestContext_cancellationDuringCleanup(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Sequential()
	ctxVar := let.Context(s)
	var ctx context.Context
	s.Test("", func(t *testcase.T) {
		ctx = ctxVar.Get(t)
		t.Must.NoError(ctx.Err())
	})
	s.Finish()
	assert.NotNil(t, ctx)
	assert.ErrorIs(t, context.Canceled, ctx.Err())
}

func TestContextWithCancel(t *testing.T) {
	s := testcase.NewSpec(t)
	ctxVar, cancelVar := let.ContextWithCancel(s)
	s.Test("", func(t *testcase.T) {
		assert.NoError(t, ctxVar.Get(t).Err())
		cancelVar.Get(t)()
		assert.ErrorIs(t, ctxVar.Get(t).Err(), context.Canceled)
	})
}

func TestPerson_smoke(t *testing.T) {
	s := testcase.NewSpec(t)

	fn := let.FirstName(s)
	ln := let.LastName(s)
	mfn := let.FirstName(s, sextype.Male)
	em := let.Email(s)

	s.Test("", func(t *testcase.T) {
		t.Must.NotEmpty(fn.Get(t))
		t.Must.NotEmpty(ln.Get(t))
		t.Must.NotEmpty(mfn.Get(t))
		t.Must.NotEmpty(em.Get(t))
		t.Eventually(func(it *testcase.T) {
			it.Must.Equal(t.Random.Contact(sextype.Male).FirstName, mfn.Get(t))
		})
	})
}

func ExampleVar() {
	var t *testcase.T
	s := testcase.NewSpec(t)

	v1 := let.Var(s, func(t *testcase.T) int {
		return t.Random.IntB(1, 7)
	})

	s.Test("", func(t *testcase.T) {
		v1.Get(t) // the random value
		v1.Get(t) // the same random value
	})
}

func ExampleVar2() {
	var t *testcase.T
	s := testcase.NewSpec(t)

	v1, v2 := let.Var2(s, func(t *testcase.T) (int, string) {
		return t.Random.IntB(1, 7), t.Random.String()
	})

	s.Test("", func(t *testcase.T) {
		v1.Get(t) // the random value constructed in the init block of Var
		v1.Get(t) // the same random value

		v2.Get(t) // the random string we made in the init block
		v2.Get(t) // the same value
	})
}

func ExampleVar3() {
	var t *testcase.T
	s := testcase.NewSpec(t)

	v1, v2, v3 := let.Var3(s, func(t *testcase.T) (int, string, bool) {
		return t.Random.IntB(1, 7), t.Random.String(), t.Random.Bool()
	})

	s.Test("", func(t *testcase.T) {
		v1.Get(t) // the random value constructed in the init block of Var
		v1.Get(t) // the same random value

		v2.Get(t) // the random string we made in the init block
		v2.Get(t) // the same value

		v3.Get(t) // the random string we made in the init block
		v3.Get(t) // the same value
	})
}

func TestVar(t *testing.T) {
	s := testcase.NewSpec(t)

	v1 := let.Var(s, func(t *testcase.T) int {
		return t.Random.IntB(1, 7)
	})

	v2, v3 := let.Var2(s, func(t *testcase.T) (string, float32) {
		return t.Random.String(), t.Random.Float32() + 1
	})

	v4, v5, v6 := let.Var3(s, func(t *testcase.T) (int, string, error) {
		return t.Random.IntB(1, 5), t.Random.String(), t.Random.Error()
	})

	s.Test("includes the current file", func(t *testcase.T) {
		file := getCurrentFileName(t)
		assert.Should(t).Contains(v1.ID, file)
		assert.Should(t).Contains(v2.ID, file)
		assert.Should(t).Contains(v3.ID, file)
		assert.Should(t).Contains(v4.ID, file)
		assert.Should(t).Contains(v5.ID, file)
		assert.Should(t).Contains(v6.ID, file)
	})

	s.Test("doesn't include the file location where the helper is defined", func(t *testcase.T) {
		assert.Should(t).NotContains(v1.ID, "let.go")
		assert.Should(t).NotContains(v2.ID, "let.go")
		assert.Should(t).NotContains(v3.ID, "let.go")
		assert.Should(t).NotContains(v4.ID, "let.go")
		assert.Should(t).NotContains(v5.ID, "let.go")
		assert.Should(t).NotContains(v6.ID, "let.go")
	})

	s.Test("variable names are all unique", func(t *testcase.T) {
		assert.Unique(t, []testcase.VarID{v1.ID, v2.ID, v3.ID, v4.ID, v5.ID, v6.ID})
	})

	s.Test("testcase.Var value retrieve works as expected", func(t *testcase.T) {
		assert.NotEmpty(t, v1.Get(t))
		assert.Equal(t, v1.Get(t), v1.Get(t))
		assert.NotEmpty(t, v2.Get(t))
		assert.Equal(t, v2.Get(t), v2.Get(t))
		assert.NotEmpty(t, v3.Get(t))
		assert.Equal(t, v3.Get(t), v3.Get(t))
		assert.NotEmpty(t, v4.Get(t))
		assert.Equal(t, v4.Get(t), v4.Get(t))
		assert.NotEmpty(t, v5.Get(t))
		assert.Equal(t, v5.Get(t), v5.Get(t))
		assert.NotEmpty(t, v6.Get(t))
		assert.Equal(t, v6.Get(t), v6.Get(t))
	})
}

func getCurrentFileName(tb testing.TB) string {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	assert.NotNil(tb, fn)
	file, _ := fn.FileLine(pc)
	return file
}

func TestHTTPTestResponseRecorder(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		handler = let.Var(s, func(t *testcase.T) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			}
		})
		response = let.HTTPTestResponseRecorder(s)
		request  = let.Var(s, func(t *testcase.T) *http.Request {
			return httptest.NewRequest(http.MethodPost, "/", nil)
		})
	)
	act := func(t *testcase.T) {
		handler.Get(t).ServeHTTP(response.Get(t), request.Get(t))
	}

	s.Then("teapot status", func(t *testcase.T) {
		act(t)

		assert.Equal(t, http.StatusTeapot, response.Get(t).Code)
	})
}

func ExampleAct() {
	// in production code
	var MyFunc = func(n int) bool {
		return n%2 == 0
	}

	// TestMyFunc(t *testing.T)
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		n = let.Int(s)
	)
	act := let.Act(func(t *testcase.T) bool {
		return MyFunc(n.Get(t))
	})

	s.Then("...", func(t *testcase.T) {
		var got bool = act(t)
		_ = got // assert
	})
}

func TestAct(tt *testing.T) {
	t := testcase.NewT(tt)
	exp := t.Random.Int()
	assert.Equal(t, exp, let.Act(func(t *testcase.T) int { return exp })(t))
}

func ExampleAct1() {
	// in production code
	var MyFunc = func(n int) bool {
		return n%2 == 0
	}

	// TestMyFunc(t *testing.T)
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		n = let.Int(s)
	)
	act := let.Act1(func(t *testcase.T) bool {
		return MyFunc(n.Get(t))
	})

	s.Then("...", func(t *testcase.T) {
		var got bool = act(t)
		_ = got // assert
	})
}

func TestAct1(tt *testing.T) {
	t := testcase.NewT(tt)
	exp := t.Random.Int()
	assert.Equal(t, exp, let.Act1(func(t *testcase.T) int { return exp })(t))
}

func ExampleAct0() {
	// in production code
	var MyFunc = func(n int) {
		// something something...
	}

	// TestMyFunc(t *testing.T)
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		n = let.Int(s)
	)
	act := let.Act0(func(t *testcase.T) {
		MyFunc(n.Get(t))
	})

	s.Then("...", func(t *testcase.T) {
		act(t) // act

		// assert outcome
	})
}

func TestAct0(tt *testing.T) {
	t := testcase.NewT(tt)
	var done bool
	act := let.Act0(func(t *testcase.T) { done = true })
	assert.NotPanic(t, func() { act(t) })
	assert.True(t, done)
}

func ExampleAct2() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		str = let.StringNC(s, 3, random.CharsetDigit())
	)
	act := let.Act2(func(t *testcase.T) (int, error) {
		return strconv.Atoi(str.Get(t))
	})

	s.Then("...", func(t *testcase.T) {
		_, _ = act(t)
	})
}

func TestAct2(tt *testing.T) {
	t := testcase.NewT(tt)
	exp1 := t.Random.Int()
	exp2 := t.Random.String()
	got1, got2 := let.Act2(func(t *testcase.T) (int, string) { return exp1, exp2 })(t)
	assert.Equal(t, exp1, got1)
	assert.Equal(t, exp2, got2)
}

func ExampleAct3() {
	// package mypkg
	var MyFunc = func(str string) (string, int, error) {
		n, err := strconv.Atoi(str)
		return str, n, err
	}

	// package mypkg_test
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		str = let.StringNC(s, 3, random.CharsetDigit())
	)
	act := let.Act3(func(t *testcase.T) (string, int, error) {
		return MyFunc(str.Get(t))
	})

	s.Then("...", func(t *testcase.T) {
		_, _, _ = act(t)
	})
}

func TestAct3(tt *testing.T) {
	t := testcase.NewT(tt)
	exp1 := t.Random.Int()
	exp2 := t.Random.String()
	exp3 := t.Random.Float64()
	got1, got2, got3 := let.Act3(func(t *testcase.T) (int, string, float64) { return exp1, exp2, exp3 })(t)
	assert.Equal(t, exp1, got1)
	assert.Equal(t, exp2, got2)
	assert.Equal(t, exp3, got3)
}

func TestVarOf_smoke(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})

	s := testcase.NewSpec(t)

	s.Context("primitive", func(s *testcase.Spec) {
		s.Context("int", func(s *testcase.Spec) {
			expInt := rnd.Int()
			vInt := let.VarOf(s, expInt)
			s.Test("smoke", func(t *testcase.T) {
				assert.NotEmpty(t, vInt.ID)
				assert.Equal(t, expInt, vInt.Get(t))
			})
		})
		s.Context("string", func(s *testcase.Spec) {
			expString := rnd.String()
			vString := let.VarOf(s, expString)
			s.Test("smoke", func(t *testcase.T) {
				assert.NotEmpty(t, vString.ID)
				assert.Equal(t, expString, vString.Get(t))
			})
		})
	})

	s.Context("struct", func(s *testcase.Spec) {
		type T struct{ V string }
		expStruct := T{V: rnd.Domain()}
		vStruct := let.VarOf(s, expStruct)
		s.Test("smoke", func(t *testcase.T) {
			assert.NotEmpty(t, vStruct.ID)
			assert.Equal(t, expStruct, vStruct.Get(t))
		})
	})

	type Mutable struct {
		V  *int
		VS []int
	}

	s.Context("nil mutable", func(s *testcase.Spec) {
		vSlice := let.VarOf[[]int](s, nil)
		vPtr := let.VarOf[*string](s, nil)

		s.Test("nil slice", func(t *testcase.T) {
			assert.NotEmpty(t, vSlice.ID)
			assert.Nil(t, vSlice.Get(t))
		})

		s.Test("nil ptr", func(t *testcase.T) {
			assert.NotEmpty(t, vPtr.ID)
			assert.Nil(t, vPtr.Get(t))
		})
	})

	s.Test("non nil mutable", func(t *testcase.T) {
		var willFail = func(t *testcase.T, fn func(s *testcase.Spec)) {
			dtb := &testcase.FakeTB{}
			spec := testcase.NewSpec(dtb)
			testcase.Sandbox(func() { fn(spec) })
			spec.Finish()
			assert.True(t, dtb.IsFailed)
		}
		willFail(t, func(s *testcase.Spec) {
			n := 42
			let.VarOf(s, &n)
		})
		willFail(t, func(s *testcase.Spec) {
			let.VarOf(s, []int{1, 2, 3})
		})
		willFail(t, func(s *testcase.Spec) {
			let.VarOf(s, map[string]int{})
		})
		willFail(t, func(s *testcase.Spec) {
			n := 42
			let.VarOf(s, Mutable{V: &n, VS: []int{n}})
		})
	})
}

func ExamplePhaser() {
	var t testing.TB
	s := testcase.NewSpec(t)

	phaser := let.Phaser(s)
	done := let.VarOf(s, false)

	s.Before(func(t *testcase.T) {
		// some background process that we want to fail until the test is ready for it to proceed forward
		go func() {
			phaser.Get(t).Wait()
			done.Set(t, true)
		}()
	})

	s.Test("smoke", func(t *testcase.T) {
		for i := 0; i < 42; i++ {
			assert.False(t, done.Get(t))
		}

		phaser.Get(t).Release() // signal that waiting goroutines can continue

		t.Eventually(func(t *testcase.T) { // assert expected outcome
			assert.True(t, done.Get(t))
		})
	})
}

func TestPhaser(t *testing.T) {
	s := testcase.NewSpec(t)

	phaser := let.Phaser(s)

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
