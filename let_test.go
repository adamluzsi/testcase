package testcase_test

import (
	"context"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"testing"
	"time"

	"go.llib.dev/testcase/pp"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/httpspec"
	"go.llib.dev/testcase/internal/caller"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/sandbox"
)

func TestLetandLetValue_returnsVar(t *testing.T) {
	s := testcase.NewSpec(t)
	v1 := testcase.Let(s, func(t *testcase.T) string {
		return t.Random.StringN(5)
	})

	s.Context(``, func(s *testcase.Spec) {
		v2 := testcase.Let(s, func(t *testcase.T) string {
			return t.Random.StringN(6)
		})

		s.Test(``, func(t *testcase.T) {
			t.Must.NotEqual(v1.ID, v2.ID)
			t.Must.NotEmpty(v1.Get(t))
			t.Must.NotEmpty(v2.Get(t))
			t.Must.Equal(v1.Get(t), v1.Get(t), "getting the same testcase.Var value must always yield the same result")
			t.Must.NotEqual(v1.Get(t), v2.Get(t))
		})
	})

	s.Test(``, func(t *testcase.T) {
		t.Must.NotEmpty(v1.Get(t))
	})
}

func TestLet_removingPreviousDeclaration(t *testing.T) {
	unsupported(t)

	dtb := &doubles.TB{}
	s := testcase.NewSpec(dtb)

	v1 := testcase.Let(s, func(t *testcase.T) string {
		return t.Random.StringN(5)
	})

	s.Context(``, func(s *testcase.Spec) {
		v1.Let(s, nil)

		s.Test(``, func(t *testcase.T) {
			v1.Get(t)
		})
	})

	pp.PP(sandbox.Run(s.Finish))
	assert.True(t, dtb.IsFailed)
}

func TestLetandLetValue_declerationInLoop_returnsUniqueVariables(t *testing.T) {
	s := testcase.NewSpec(t)

	var letValues []testcase.Var[int]
	var lets []testcase.Var[int]
	for i := 0; i < 42; i++ {
		i := i
		letValues = append(letValues, testcase.LetValue(s, i))
		lets = append(lets, testcase.Let(s, func(t *testcase.T) int { return i }))
	}

	s.Test(``, func(t *testcase.T) {
		for i := 0; i < 42; i++ {
			t.Must.Equal(i, letValues[i].Get(t))
			t.Must.Equal(i, lets[i].Get(t))
		}
	})
}

func TestLetandLetValue_declerationInHelper_returnsUniqueVariables(t *testing.T) {
	s := testcase.NewSpec(t)

	var letValues []testcase.Var[int]
	var lets []testcase.Var[int]
	for i := 0; i < 42; i++ {
		i := i
		letValues = append(letValues, letValueForTestLetandLetValueDeclerationInHelper(s, i))
		lets = append(lets, letForTestLetandLetValueDeclerationInHelper(s, i))
	}

	s.Test(``, func(t *testcase.T) {
		for i := 0; i < 42; i++ {
			t.Must.Equal(i, letValues[i].Get(t))
			t.Must.Equal(i, lets[i].Get(t))
		}
	})
}

func letForTestLetandLetValueDeclerationInHelper(s *testcase.Spec, i int) testcase.Var[int] {
	return testcase.Let(s, func(t *testcase.T) int {
		return i
	})
}

func letValueForTestLetandLetValueDeclerationInHelper(s *testcase.Spec, i int) testcase.Var[int] {
	return testcase.Let(s, func(t *testcase.T) int {
		return i
	})
}

func TestLetValue_returnsVar(t *testing.T) {
	s := testcase.NewSpec(t)
	counter := testcase.LetValue(s, 0)

	s.Test(``, func(t *testcase.T) {
		t.Must.Equal(0, counter.Get(t))
		counter.Set(t, 1)
		t.Must.Equal(1, counter.Get(t))
		counter.Set(t, 2)
	})
}

func TestLet_posName(t *testing.T) {
	t.Run("multiple Let in for", func(t *testing.T) {
		s := testcase.NewSpec(t)
		var lets []testcase.Var[int]
		for i := 0; i < 2; i++ {
			i := i
			lets = append(lets, testcase.Let(s, func(t *testcase.T) int { return i }))
		}
		s.Test(``, func(t *testcase.T) {
			v := lets[len(lets)-1]
			t.Must.Contains(v.ID, "let_test.go")
			t.Must.Contains(v.ID, "#[1]")
		})
	})

	letInt := func(s *testcase.Spec, v int) testcase.Var[int] {
		return testcase.LetValue(s, v)
	}
	t.Run("multiple let from helper func across different spec contexts", func(t *testing.T) {
		s := testcase.NewSpec(t)
		a := letInt(s, 1)
		s.Context("sub", func(s *testcase.Spec) {
			b := letInt(s, 2)

			s.Test("test", func(t *testcase.T) {
				t.Must.Equal(a.Get(t), 1)
				t.Must.Equal(b.Get(t), 2)
			})
		})
	})
}

func TestLet_withNilBlock(tt *testing.T) {
	stub := &doubles.TB{}
	defer stub.Finish()
	s := testcase.NewSpec(stub)
	v := testcase.Let[int](s, nil)
	var ran bool
	s.Test("", func(t *testcase.T) {
		ran = true
		out := sandbox.Run(func() { v.Get(t) })
		assert.False(t, out.OK)
	})
	s.Finish()
	assert.True(tt, ran)
	logs := stub.Logs.String()
	assert.Contains(tt, logs, "is not found")
	assert.Contains(tt, logs, "Did you mean?")
}

func TestLetValue_withNil(tt *testing.T) {
	it := assert.MakeIt(tt)
	stub := &doubles.TB{}
	defer stub.Finish()
	s := testcase.NewSpec(stub)
	v := testcase.Let[[]int](s, nil)
	out := sandbox.Run(func() { v.LetValue(s, nil) })
	tt.Log(stub.Logs.String())
	assert.True(tt, out.OK)
	assert.False(tt, stub.Failed())
	var ran bool
	s.Test("", func(t *testcase.T) {
		ran = true
		out := sandbox.Run(func() { it.Must.Nil(v.Get(t)) })
		assert.True(t, out.OK)
	})
	s.Finish()
	assert.True(tt, ran)
	assert.False(tt, stub.IsFailed)
	assert.False(tt, stub.IsSkipped)
}

func TestLet_varID_testFile(t *testing.T) {
	var frame runtime.Frame
	caller.Until(caller.NonTestCaseFrame, func(f runtime.Frame) bool {
		frame = f
		return true
	})

	s := testcase.NewSpec(t)
	v := testcase.Let[int](s, nil)
	assert.Contains(t, v.ID, "_test.go")
	assert.Contains(t, v.ID, filepath.Base(frame.File))
}

func TestLetValue_varID_testFile(t *testing.T) {
	var frame runtime.Frame
	caller.Until(caller.NonTestCaseFrame, func(f runtime.Frame) bool {
		frame = f
		return true
	})

	s := testcase.NewSpec(t)
	v := testcase.LetValue(s, 42)
	assert.Contains(t, v.ID, "_test.go")
	assert.Contains(t, v.ID, filepath.Base(frame.File))
}

func TestLet_letVarIDInNonCoreTestcasePackage(t *testing.T) {
	var frame runtime.Frame
	caller.Until(caller.NonTestCaseFrame, func(f runtime.Frame) bool {
		frame = f
		return true
	})

	s := testcase.NewSpec(t)
	resp := httpspec.LetResponseRecorder(s)
	t.Logf("id: %s", resp.ID)
	assert.NotContains(t, resp.ID, "_test.go")
	assert.NotContains(t, resp.ID, filepath.Base(frame.File))
	assert.Contains(t, resp.ID, filepath.Dir(frame.File))
}

func ExampleRegisterImmutableType() {
	type T struct{}

	testcase.RegisterImmutableType[T]()

	s := testcase.NewSpec((testing.TB)(nil))
	v := testcase.LetValue(s, T{})
	_ = v

	s.Test("", func(t *testcase.T) {})
}

func TestLetValue_struct(t *testing.T) {
	t.Run("with immutable fields only", func(t *testing.T) {
		type Sub struct {
			V string
		}
		type T struct {
			A string
			B int
			C Sub
		}
		s := testcase.NewSpec(t)
		s.HasSideEffect()
		v := testcase.LetValue(s, T{
			A: "The Answer",
			B: 42,
		})
		s.Test("", func(t *testcase.T) {
			t.Must.Equal("The Answer", v.Get(t).A)
			t.Must.Equal(42, v.Get(t).B)
		})
	})
	t.Run("with mutable fields", func(t *testing.T) {
		type Sub struct {
			V *string
		}
		type T1 struct{ V *string }
		type T2 struct{ V Sub }
		type T3 struct{ Vs []string }
		type T4 struct{ KVs map[string]string }
		type T5 struct{ Vs [5]int }
		type T6 struct{ VChan chan int }

		var fail = func(t *testing.T, fn func(s *testcase.Spec)) {
			ftb := &testcase.FakeTB{}
			s := testcase.NewSpec(ftb)
			out := testcase.Sandbox(func() {
				fn(s)
			})
			assert.True(t, ftb.IsFailed)
			assert.False(t, out.OK)
		}

		fail(t, func(s *testcase.Spec) {
			testcase.LetValue(s, T1{})
		})
		fail(t, func(s *testcase.Spec) {
			testcase.LetValue(s, T2{})
		})
		fail(t, func(s *testcase.Spec) {
			testcase.LetValue(s, T3{})
		})
		fail(t, func(s *testcase.Spec) {
			testcase.LetValue(s, T4{})
		})
		fail(t, func(s *testcase.Spec) {
			testcase.LetValue(s, T5{})
		})
		fail(t, func(s *testcase.Spec) {
			testcase.LetValue(s, T6{})
		})

		t.Run("but registered as an exception", func(t *testing.T) {
			t.Cleanup(testcase.RegisterImmutableType[T1]())
			t.Cleanup(testcase.RegisterImmutableType[T2]())
			t.Cleanup(testcase.RegisterImmutableType[T3]())
			td := testcase.RegisterImmutableType[T4]()
			ftb := &testcase.FakeTB{}
			s := testcase.NewSpec(ftb)
			testcase.LetValue(s, T1{})
			testcase.LetValue(s, T2{})
			testcase.LetValue(s, T3{})
			testcase.LetValue(s, T4{})
			assert.False(t, ftb.IsFailed)
			td()
			fail(t, func(s *testcase.Spec) {
				testcase.LetValue(s, T4{})
			})
		})
	})
}

func TestLetValue_stdlibExceptions(t *testing.T) {
	s := testcase.NewSpec(t)
	*time.Now().Location() = *time.UTC
	testcase.LetValue(s, time.Now())
	testcase.LetValue(s, *time.Local)
	testcase.LetValue(s, *time.UTC)
	testcase.LetValue(s, context.Background())
	testcase.LetValue(s, context.TODO())
	testcase.LetValue(s, reflect.TypeOf(42))
	testcase.LetValue(s, *regexp.MustCompile(`^hello$`))
}

func TestLet2(t *testing.T) {
	t.Run("tuple creation possible and single value retrieve works", func(t *testing.T) {
		s := testcase.NewSpec(t)
		defer s.Finish()

		v, b := testcase.Let2(s, func(t *testcase.T) (int, string) {
			return t.Random.Int(), t.Random.String()
		})

		s.Test("", func(t *testcase.T) {
			var (
				vv int
				bv string
			)
			t.Must.Within(time.Second, func(context.Context) {
				vv = v.Get(t)
				bv = b.Get(t)
			})
			t.Must.NotEmpty(vv)
			t.Must.NotEmpty(bv)
			t.Random.Repeat(2, 5, func() {
				t.Must.Equal(v.Get(t), vv)
				t.Must.Equal(b.Get(t), bv)
			})
		})
	})

	t.Run("value is not initialised before calling it", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()

		var isCalled bool
		v, b := testcase.Let2(s, func(t *testcase.T) (int, string) {
			isCalled = true
			return t.Random.Int(), t.Random.String()
		})
		_, _ = v, b

		s.Test("", func(t *testcase.T) {})
		s.Finish()

		assert.False(t, isCalled)
	})
}

func TestLet3(t *testing.T) {
	t.Run("tuple creation possible and single value retrieve works", func(t *testing.T) {
		s := testcase.NewSpec(t)
		defer s.Finish()

		v, b, n := testcase.Let3(s, func(t *testcase.T) (int, string, float32) {
			return t.Random.Int(), t.Random.String(), t.Random.Float32()
		})

		s.Test("", func(t *testcase.T) {
			var (
				vv int
				bv string
				nv float32
			)
			t.Must.Within(time.Second, func(context.Context) {
				vv = v.Get(t)
				bv = b.Get(t)
				nv = n.Get(t)
			})
			t.Must.NotEmpty(vv)
			t.Must.NotEmpty(bv)
			t.Must.NotEmpty(nv)
			t.Random.Repeat(2, 5, func() {
				t.Must.Equal(v.Get(t), vv)
				t.Must.Equal(b.Get(t), bv)
				t.Must.Equal(n.Get(t), nv)
			})
		})
	})

	t.Run("value is not initialised before calling it", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()

		var isCalled bool
		v, b, n := testcase.Let3(s, func(t *testcase.T) (int, string, bool) {
			isCalled = true
			return t.Random.Int(), t.Random.String(), t.Random.Bool()
		})
		_, _, _ = v, b, n

		s.Test("", func(t *testcase.T) {})
		s.Finish()

		assert.False(t, isCalled)
	})
}
