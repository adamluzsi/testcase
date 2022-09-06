package testcase

import (
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/doubles"
	"testing"
)

var (
	_ testingT = &testing.T{}
	_ testingB = &testing.B{}
)

func TestToSpec_smoke(t *testing.T) {
	assertToSpec := func(tb testing.TB, s *Spec) {
		tb.Helper()
		var ran bool
		v := LetValue(s, 42)
		s.HasSideEffect()
		s.Test("", func(t *T) {
			assert.Equal(tb, 42, v.Get(t))
			ran = true
		})
		s.Finish()
		assert.True(tb, ran)
	}

	assertToSpec(t, ToSpec(NewSpec(t)))
	assertToSpec(t, ToSpec(t))
	assertToSpec(t, ToSpec(NewT(t, nil)))

	dtb := &doubles.TB{}
	assertToSpec(t, ToSpec(dtb))
	dtb.Finish()

	var tb testing.TB = t
	assertToSpec(t, ToSpec(&tb))
}

func BenchmarkTestToSpec(b *testing.B) {
	s := ToSpec(b)
	var ran bool
	v := LetValue(s, 42)
	s.HasSideEffect()
	s.Test("", func(t *T) {
		assert.Equal(b, 42, v.Get(t))
		ran = true
		t.Skip("TEST")
	})
	s.Finish()
	assert.True(b, ran)
	b.Skip("TEST")
}

func TestToT(t *testing.T) {

}