package internal

import "testing"

var CacheEnabled = true

func DisableCache(tb testing.TB) {
	og := CacheEnabled
	tb.Cleanup(func() { CacheEnabled = og })
	CacheEnabled = false
}
