package slicekit_test

import (
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/slicekit"
)

func ExampleLookup() {
	vs := []int{2, 4, 8, 16}
	slicekit.Lookup(vs, 0)      // -> return 2, true
	slicekit.Lookup(vs, 0-1)    // lookup previous -> return 0, false
	slicekit.Lookup(vs, 0+1)    // lookup next -> return 4, true
	slicekit.Lookup(vs, 0+1000) // lookup 1000th element -> return 0, false
}

func TestLookup_smoke(t *testing.T) {
	vs := []int{2, 4, 8, 16}

	v, ok := slicekit.Lookup(vs, 0)
	assert.Equal(t, ok, true)
	assert.Equal(t, v, 2)

	v, ok = slicekit.Lookup(vs, -1)
	assert.Equal(t, ok, true)
	assert.Equal(t, v, 16)

	v, ok = slicekit.Lookup(vs, 0+1)
	assert.Equal(t, ok, true)
	assert.Equal(t, v, 4)

	v, ok = slicekit.Lookup(vs, 0+1000)
	assert.Equal(t, ok, false)
	assert.Equal(t, v, 0)

	v, ok = slicekit.Lookup(vs, 0+1000)
	assert.Equal(t, ok, false)
	assert.Equal(t, v, 0)

	for i, exp := range vs {
		got, ok := slicekit.Lookup(vs, i)
		assert.Equal(t, ok, true)
		assert.Equal(t, exp, got)
	}
}

func TestLookup_negativeIndex(t *testing.T) {
	vs := []int{2, 4, 8, 16, 32}

	v, ok := slicekit.Lookup(vs, -1)
	assert.Equal(t, ok, true)
	assert.Equal(t, v, 32)

	v, ok = slicekit.Lookup(vs, -2)
	assert.Equal(t, ok, true)
	assert.Equal(t, v, 16)

	v, ok = slicekit.Lookup(vs, -3)
	assert.Equal(t, ok, true)
	assert.Equal(t, v, 8)

	v, ok = slicekit.Lookup(vs, (len(vs)+1)*-1)
	assert.Equal(t, ok, false)
	assert.Empty(t, v)
}

func TestReverseLookup(t *testing.T) {
	vs := []int{2, 4, 8, 16, 32}

	v, ok := slicekit.ReverseLookup(vs, 0)
	assert.Equal(t, ok, true)
	assert.Equal(t, v, 32)

	v, ok = slicekit.ReverseLookup(vs, 1)
	assert.Equal(t, ok, true)
	assert.Equal(t, v, 16)

	v, ok = slicekit.ReverseLookup(vs, 2)
	assert.Equal(t, ok, true)
	assert.Equal(t, v, 8)

	v, ok = slicekit.ReverseLookup(vs, 3)
	assert.Equal(t, ok, true)
	assert.Equal(t, v, 4)

	v, ok = slicekit.ReverseLookup(vs, 4)
	assert.Equal(t, ok, true)
	assert.Equal(t, v, 2)

	v, ok = slicekit.ReverseLookup(vs, 5)
	assert.Equal(t, ok, false)
	assert.Empty(t, v)
}

func ExampleMerge() {
	var (
		a   = []string{"a", "b", "c"}
		b   = []string{"1", "2", "3"}
		c   = []string{"1", "B", "3"}
		out = slicekit.Merge(a, b, c)
	)
	_ = out // []string{"a", "b", "c", "1", "2", "3", "1", "B", "3"}
}

func TestMerge(t *testing.T) {
	t.Run("all slice merged into one", func(t *testing.T) {
		var (
			a   = []string{"a", "b", "c"}
			b   = []string{"1", "2", "3"}
			c   = []string{"1", "B", "3"}
			out = slicekit.Merge(a, b, c)
		)
		assert.Equal(t, out, []string{
			"a", "b", "c",
			"1", "2", "3",
			"1", "B", "3",
		})
	})
	t.Run("input slices are not affected by the merging process", func(t *testing.T) {
		var (
			a = []string{"a", "b", "c"}
			b = []string{"1", "2", "3"}
			c = []string{"1", "B", "3"}
			_ = slicekit.Merge(a, b, c)
		)
		assert.Equal(t, a, []string{"a", "b", "c"})
		assert.Equal(t, b, []string{"1", "2", "3"})
		assert.Equal(t, c, []string{"1", "B", "3"})
	})
}
