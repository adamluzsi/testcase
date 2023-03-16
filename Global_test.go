package testcase

import (
	"math/rand"
	"testing"
)

func TestGlobal_Before(t *testing.T) {
	t.Cleanup(func() { Global = global{} })
	v := Var[int]{
		ID: "v",
		Init: func(t *T) int {
			return 0
		},
	}
	n1 := rand.Intn(10)
	Global.Before(func(t *T) {
		v.Set(t, v.Get(t)+n1)
	})
	n2 := rand.Intn(10)
	Global.Before(func(t *T) {
		v.Set(t, v.Get(t)+n2)
	})
	for i := 0; i < 42; i++ {
		s := NewSpec(t)
		s.Test("", func(t *T) { t.Must.Equal(n1+n2, v.Get(t)) })
	}
}

func TestGlobal_Before_race(t *testing.T) {
	t.Cleanup(func() { Global = global{} })
	Race(func() {
		Global.Before(func(t *T) {})
	}, func() {
		Global.Before(func(t *T) {})
	}, func() {
		Global.Before(func(t *T) {})
	})
}
