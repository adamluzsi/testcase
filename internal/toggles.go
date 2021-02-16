package internal

import (
	"testing"
)

func SetupCacheFlush(tb testing.TB) {
	CacheFlush()
	tb.Cleanup(CacheFlush)
}

var cacheFlushFns []func()

func RegisterCacheFlush(fn func()) struct{} {
	cacheFlushFns = append(cacheFlushFns, fn)
	return struct{}{}
}

func CacheFlush() {
	for _, fn := range cacheFlushFns {
		fn()
	}
}
