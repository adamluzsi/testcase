package assert

import (
	"context"
	"io"
	"testing"
	"time"
)

func True(tb testing.TB, v bool, msg ...Message) {
	tb.Helper()
	Must(tb).True(v, msg...)
}

func False(tb testing.TB, v bool, msg ...Message) {
	tb.Helper()
	Must(tb).False(v, msg...)
}

func Nil(tb testing.TB, v any, msg ...Message) {
	tb.Helper()
	Must(tb).Nil(v, msg...)
}

func NotNil(tb testing.TB, v any, msg ...Message) {
	tb.Helper()
	Must(tb).NotNil(v, msg...)
}

func Empty(tb testing.TB, v any, msg ...Message) {
	tb.Helper()
	Must(tb).Empty(v, msg...)
}

func NotEmpty(tb testing.TB, v any, msg ...Message) {
	tb.Helper()
	Must(tb).NotEmpty(v, msg...)
}

func Panic(tb testing.TB, blk func(), msg ...Message) any {
	tb.Helper()
	return Must(tb).Panic(blk, msg...)
}

func NotPanic(tb testing.TB, blk func(), msg ...Message) {
	tb.Helper()
	Must(tb).NotPanic(blk, msg...)
}

func Equal[T any](tb testing.TB, v, oth T, msg ...Message) {
	tb.Helper()
	Must(tb).Equal(v, oth, msg...)
}

func NotEqual[T any](tb testing.TB, v, oth T, msg ...Message) {
	tb.Helper()
	Must(tb).NotEqual(v, oth, msg...)
}

func Contain(tb testing.TB, haystack, needle any, msg ...Message) {
	tb.Helper()
	Must(tb).Contain(haystack, needle, msg...)
}

func NotContain(tb testing.TB, haystack, v any, msg ...Message) {
	tb.Helper()
	Must(tb).NotContain(haystack, v, msg...)
}

func ContainExactly[T any /* Map or Slice */](tb testing.TB, v, oth T, msg ...Message) {
	tb.Helper()
	Must(tb).ContainExactly(v, oth, msg...)
}

func Sub[T any](tb testing.TB, haystack, needle []T, msg ...Message) {
	tb.Helper()
	Must(tb).Sub(haystack, needle, msg...)
}

func ErrorIs(tb testing.TB, err, oth error, msg ...Message) {
	tb.Helper()
	Must(tb).ErrorIs(err, oth, msg...)
}

func Error(tb testing.TB, err error, msg ...Message) {
	tb.Helper()
	Must(tb).Error(err, msg...)
}

func NoError(tb testing.TB, err error, msg ...Message) {
	tb.Helper()
	Must(tb).NoError(err, msg...)
}

func Read[T string | []byte](tb testing.TB, v T, r io.Reader, msg ...Message) {
	tb.Helper()
	Must(tb).Read(v, r, msg...)
}

func ReadAll(tb testing.TB, r io.Reader, msg ...Message) []byte {
	tb.Helper()
	return Must(tb).ReadAll(r, msg...)
}

func Within(tb testing.TB, timeout time.Duration, blk func(context.Context), msg ...Message) {
	tb.Helper()
	Must(tb).Within(timeout, blk, msg...)
	// Returning *Async here doesn’t make sense because if the assertion fails,
	// FailNow will terminate the current goroutine regardless.
}

func NotWithin(tb testing.TB, timeout time.Duration, blk func(context.Context), msg ...Message) *Async {
	tb.Helper()
	return Must(tb).NotWithin(timeout, blk, msg...)
}

func MatchRegexp[T ~string | []byte](tb testing.TB, v T, expr string, msg ...Message) {
	tb.Helper()
	Must(tb).MatchRegexp(string(v), expr, msg...)
}

func NotMatchRegexp[T ~string | []byte](tb testing.TB, v T, expr string, msg ...Message) {
	tb.Helper()
	Must(tb).NotMatchRegexp(string(v), expr, msg...)
}

func Eventually[T time.Duration | int](tb testing.TB, durationOrCount T, blk func(t It)) {
	tb.Helper()
	Must(tb).Eventually(durationOrCount, blk)
}

// AnyOf is an assertion helper that deems the test successful
// if any of the declared assertion cases pass.
// This is commonly used when multiple valid formats are acceptable
// or when working with a list where any element meeting a certain criteria is considered sufficient.
func AnyOf(tb testing.TB, blk func(a *A), msg ...Message) {
	tb.Helper()
	Must(tb).AnyOf(blk)
}

// Unique will verify if the given list has unique elements.
func Unique[T any](tb testing.TB, vs []T, msg ...Message) {
	tb.Helper()
	Must(tb).Unique(vs, msg...)
}

// NotUnique will verify if the given list has at least one duplicated element.
func NotUnique[T any](tb testing.TB, vs []T, msg ...Message) {
	tb.Helper()
	Must(tb).NotUnique(vs, msg...)
}
