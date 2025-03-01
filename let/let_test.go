package let_test

import (
	"context"
	"runtime"
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

func Test_smoke(t *testing.T) {
	s := testcase.NewSpec(t)

	Context := let.Context(s)
	Error := let.Error(s)
	String := let.String(s)
	StringNC := let.StringNC(s, 42, random.CharsetASCII())
	Bool := let.Bool(s)
	Int := let.Int(s)
	IntN := let.IntN(s, 42)
	IntB := let.IntB(s, 7, 42)
	Time := let.Time(s)
	TimeB := let.TimeB(s, time.Now().AddDate(-1, 0, 0), time.Now())
	UUID := let.UUID(s)
	Element := let.ElementFrom[string](s, "foo", "bar", "baz")
	DurationBetween := let.DurationBetween(s, time.Second, time.Minute)

	charsterIs := func(t *testcase.T, cs, str string) {
		for _, v := range str {
			t.Must.Contain(cs, string(v))
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
		t.Eventually(func(t *testcase.T) {
			t.Must.NotEmpty(IntN.Get(testcase.ToT(&t.TB)))
		})
		t.Must.NotEmpty(IntB.Get(t))
		t.Must.NotEmpty(DurationBetween.Get(t))
		t.Must.True(time.Second <= DurationBetween.Get(t))
		t.Must.True(DurationBetween.Get(t) <= time.Minute)
		t.Must.NotEmpty(Time.Get(t))
		t.Must.NotEmpty(TimeB.Get(t))
		t.Must.True(TimeB.Get(t).After(time.Now().AddDate(-1, 0, -1)))
		t.Must.NotEmpty(UUID.Get(t))
		t.Must.NotEmpty(Element.Get(t))
		t.Eventually(func(it *testcase.T) {
			it.Must.True(Bool.Get(testcase.ToT(&t.TB)))
		})
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
		assert.Should(t).Contain(v1.ID, file)
		assert.Should(t).Contain(v2.ID, file)
		assert.Should(t).Contain(v3.ID, file)
		assert.Should(t).Contain(v4.ID, file)
		assert.Should(t).Contain(v5.ID, file)
		assert.Should(t).Contain(v6.ID, file)
	})

	s.Test("doesn't include the file location where the helper is defined", func(t *testcase.T) {
		assert.Should(t).NotContain(v1.ID, "let.go")
		assert.Should(t).NotContain(v2.ID, "let.go")
		assert.Should(t).NotContain(v3.ID, "let.go")
		assert.Should(t).NotContain(v4.ID, "let.go")
		assert.Should(t).NotContain(v5.ID, "let.go")
		assert.Should(t).NotContain(v6.ID, "let.go")
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
