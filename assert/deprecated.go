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

// NotContain is a backward port func to enable migration to assert.NotContains
//
// Deprecated: use assert.NotContains instead of assert.NotContain
func NotContain(tb testing.TB, haystack, v any, msg ...Message) {
	tb.Helper()
	NotContains(tb, haystack, v, msg...)
}

// Contain is a backward port func to enable migration to assert.Asserter#Contains
//
// Deprecated: use assert.Asserter#Contains instead of assert.Asserter#Contain
func (a Asserter) Contain(haystack, needle any, msg ...Message) {
	a.TB.Helper()
	a.Contains(haystack, needle, msg...)
}

// NotContain is a backward port func to enable migration to assert.Asserter#NotContains
//
// Deprecated: use assert.Asserter#NotContains instead of assert.Asserter#NotContain
func (a Asserter) NotContain(haystack, needle any, msg ...Message) {
	a.TB.Helper()
	a.NotContains(haystack, needle, msg...)
}

// ContainExactly is a backward port func to enable migration to assert.ContainsExactly
//
// Deprecated: use assert.ContainsExactly instead of assert.ContainExactly
func ContainExactly[T any /* Map or Slice */](tb testing.TB, v, oth T, msg ...Message) {
	tb.Helper()
	ContainsExactly(tb, v, oth, msg...)
}

// ContainExactly is a backward port func to enable migration to assert.Asserter#ContainsExactly
//
// Deprecated: use assert.Asserter#ContainsExactly instead of assert.Asserter#ContainExactly
func (a Asserter) ContainExactly(v, oth any /* slice | map */, msg ...Message) {
	a.TB.Helper()
	a.ContainsExactly(v, oth, msg...)
}
