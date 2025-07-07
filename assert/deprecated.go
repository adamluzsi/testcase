package assert

import "testing"

// RetryStrategy
//
// Deprecated: use Loop instead
type RetryStrategy = Loop

// RetryStrategyFunc
//
// Deprecated: use LoopFunc instead
type RetryStrategyFunc = LoopFunc

// Contain is a backward port func to enable migration to assert.Contains
//
// Deprecated: use assert.Contains instead of assert.Contain
func Contain(tb testing.TB, haystack, needle any, msg ...Message) {
	tb.Helper()
	Contains(tb, haystack, needle, msg...)
}

// Contain is a backward port func to enable migration to assert.Asserter#Contains
//
// Deprecated: use assert.Asserter#Contains instead of assert.Asserter#Contain
func (a Asserter) Contain(haystack, needle any, msg ...Message) {
	a.TB.Helper()
	a.Contains(haystack, needle, msg...)
}
