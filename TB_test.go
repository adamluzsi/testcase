package testcase

import (
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
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
	assertToSpec(t, ToSpec(NewTWithSpec(t, nil)))

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
	t.Run("*testcase.T", func(t *testing.T) {
		tc := NewT(t)
		assert.Equal(t, tc, ToT(tc))
	})
	t.Run("*testing.TB", func(t *testing.T) {
		tc := NewT(t)
		var tb testing.TB = tc
		assert.Equal(t, tc, ToT(&tb))
	})
	t.Run("*testing.T", func(t *testing.T) {
		tc := ToT(t)
		assert.NotNil(t, tc)
		assert.Equal[testing.TB](t, tc.TB, t)
	})
	t.Run("*testing.F", func(t *testing.T) {
		f := &testing.F{}
		tc := ToT(f)
		assert.NotNil(t, tc)
		assert.Equal[testing.TB](t, f, tc.TB)
	})
	t.Run("*doubles.TB", func(t *testing.T) {
		dtb := &doubles.TB{}
		tc := ToT(dtb)
		assert.NotNil(t, tc)
		tc.Log("ok")
		assert.NotEmpty(t, dtb.Logs.String())
	})
	t.Run("*TBRunner", func(t *testing.T) {
		dtb := &doubles.TB{}
		var tbr TBRunner = dtb
		tc := ToT(&tbr)
		tc.Log("ok")
		assert.NotEmpty(t, dtb.Logs.String())
	})
	t.Run("assert.It", func(t *testing.T) {
		otc := NewT(t)
		var it = assert.MakeIt(otc)
		tc := ToT(it)
		assert.Equal(t, otc, tc)
	})
	t.Run("type that implements testing.TB and passed as *testing.TB", func(t *testing.T) {
		type STB struct{ testing.TB }
		dtb := &doubles.TB{}
		stb := STB{TB: dtb}
		var tb testing.TB = stb
		tc := ToT(&tb)
		tc.Log("ok")
		assert.NotEmpty(t, dtb.Logs.String())
	})
	t.Run("type that implements testing.TB and has *testcase.T used as TB", func(t *testing.T) {
		type STB struct{ testing.TB }
		tc := NewT(t)
		stb := STB{TB: tc}
		var tb testing.TB = stb
		got := ToT(&tb)
		assert.Equal(t, tc, got)
		var tb2 testing.TB = &testing.F{}
		assert.NotNil(t, ToT(&tb2))
	})
	t.Run("type that implements testing.TB usign an uninitialised testing.TB embeded field", func(t *testing.T) {
		dtb := &doubles.TB{}
		rtb := &doubles.RecorderTB{TB: dtb}
		var tb testing.TB = rtb
		tc := ToT(&tb)
		assert.Equal[testing.TB](t, tc.TB, dtb)
	})
}
