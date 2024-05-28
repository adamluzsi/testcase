package testcase_test

import (
	"context"
	"path/filepath"
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
			t.Must.Contain(v.ID, "let_test.go")
			t.Must.Contain(v.ID, "#[1]")
		})
	})

	letInt := func(s *testcase.Spec, v int) testcase.Var[int] {
		return testcase.LetValue[int](s, v)
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
	it := assert.MakeIt(tt)
	stub := &doubles.TB{}
	defer stub.Finish()
	s := testcase.NewSpec(stub)
	v := testcase.Let[int](s, nil)
	var ran bool
	s.Test("", func(t *testcase.T) {
		ran = true
		out := sandbox.Run(func() { v.Get(t) })
		it.Must.False(out.OK)
	})
	s.Finish()
	it.Must.True(ran)
	logs := stub.Logs.String()
	it.Must.Contain(logs, "is not found")
	it.Must.Contain(logs, "Did you mean?")
}

func TestLetValue_withNil(tt *testing.T) {
	it := assert.MakeIt(tt)
	stub := &doubles.TB{}
	defer stub.Finish()
	s := testcase.NewSpec(stub)
	v := testcase.Let[[]int](s, nil)
	out := sandbox.Run(func() { v.LetValue(s, nil) })
	tt.Log(stub.Logs.String())
	it.Must.True(out.OK)
	it.Must.False(stub.Failed())
	var ran bool
	s.Test("", func(t *testcase.T) {
		ran = true
		out := sandbox.Run(func() { it.Must.Nil(v.Get(t)) })
		it.Must.True(out.OK)
	})
	s.Finish()
	it.Must.True(ran)
	it.Must.False(stub.IsFailed)
	it.Must.False(stub.IsSkipped)
}

func TestLet_varID_testFile(t *testing.T) {
	var frame runtime.Frame
	caller.Until(caller.NonTestCaseFrame, func(f runtime.Frame) bool {
		frame = f
		return true
	})

	s := testcase.NewSpec(t)
	v := testcase.Let[int](s, nil)
	assert.Contain(t, v.ID, "_test.go")
	assert.Contain(t, v.ID, filepath.Base(frame.File))
}

func TestLetValue_varID_testFile(t *testing.T) {
	var frame runtime.Frame
	caller.Until(caller.NonTestCaseFrame, func(f runtime.Frame) bool {
		frame = f
		return true
	})

	s := testcase.NewSpec(t)
	v := testcase.LetValue[int](s, 42)
	assert.Contain(t, v.ID, "_test.go")
	assert.Contain(t, v.ID, filepath.Base(frame.File))
}

func TestLet_letVarIDInNonCoreTestcasePackage(t *testing.T) {
	var frame runtime.Frame
	caller.Until(caller.NonTestCaseFrame, func(f runtime.Frame) bool {
		frame = f
		return true
	})

	s := testcase.NewSpec(t)
	resp := httpspec.LetResponseRecorder(s)
	t.Logf(resp.ID)
	assert.NotContain(t, resp.ID, "_test.go")
	assert.NotContain(t, resp.ID, filepath.Base(frame.File))
	assert.Contain(t, resp.ID, filepath.Dir(frame.File))
}

func TestLetValue_struct(t *testing.T) {
	type StructWithoutMutableField struct {
		A string
		B int
	}

	s := testcase.NewSpec(t)
	s.HasSideEffect()
	v := testcase.LetValue(s, StructWithoutMutableField{
		A: "The Answer",
		B: 42,
	})

	s.Test("", func(t *testcase.T) {
		t.Must.Equal("The Answer", v.Get(t).A)
		t.Must.Equal(42, v.Get(t).B)
	})
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
