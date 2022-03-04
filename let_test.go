package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
)

func TestLetandLetValue_returnsVar(t *testing.T) {
	s := testcase.NewSpec(t)
	v1 := testcase.Let(s, func(t *testcase.T) string {
		return t.Random.String()
	})

	s.Context(``, func(s *testcase.Spec) {
		v2 := testcase.LetValue(s, fixtures.Random.String())

		s.Test(``, func(t *testcase.T) {
			t.Must.NotEqual(v1.ID, v2.ID)
			t.Must.NotEmpty(v1.Get(t))
			t.Must.NotEmpty(v2.Get(t))
			t.Must.NotEqual(v1.Get(t), v2.Get(t))
			t.Must.Equal(v1.Get(t), v1.Get(t))
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
