package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
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
