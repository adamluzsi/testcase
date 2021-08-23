package testcase_test

import (
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
)

func ExampleT_random() {
	var tb testing.TB
	s := testcase.NewSpec(tb)
	s.Test(``, func(t *testcase.T) {
		_ = t.Random.Int()
		_ = t.Random.IntBetween(0, 42)
		_ = t.Random.IntN(42)
		_ = t.Random.Float32()
		_ = t.Random.Float64()
		_ = t.Random.String()
		_ = t.Random.StringN(42)
		_ = t.Random.StringNWithCharset(42, "abc")
		_ = t.Random.Bool()
		_ = t.Random.Time()
		_ = t.Random.TimeN(time.Now(), 0, 4, 2)
		_ = t.Random.TimeBetween(time.Now().Add(-1*time.Hour), time.Now().Add(time.Hour))
		_ = t.Random.ElementFromSlice([]int{1, 2, 3}).(int)
		_ = t.Random.KeyFromMap(map[string]struct{}{`foo`: {}, `bar`: {}, `baz`: {}}).(string)
	})
}
