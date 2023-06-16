package assert

import (
	"context"
	"io"
	"testing"
	"time"
)

func True(tb testing.TB, v bool, msg ...any) {
	tb.Helper()
	Must(tb).True(v, msg...)
}

func False(tb testing.TB, v bool, msg ...any) {
	tb.Helper()
	Must(tb).False(v, msg...)
}

func Nil(tb testing.TB, v any, msg ...any) {
	tb.Helper()
	Must(tb).Nil(v, msg...)
}

func NotNil(tb testing.TB, v any, msg ...any) {
	tb.Helper()
	Must(tb).NotNil(v, msg...)
}

func Empty(tb testing.TB, v any, msg ...any) {
	tb.Helper()
	Must(tb).Empty(v, msg...)
}

func NotEmpty(tb testing.TB, v any, msg ...any) {
	tb.Helper()
	Must(tb).NotEmpty(v, msg...)
}

func Panic(tb testing.TB, blk func(), msg ...any) any {
	tb.Helper()
	return Must(tb).Panic(blk, msg...)
}

func NotPanic(tb testing.TB, blk func(), msg ...any) {
	tb.Helper()
	Must(tb).NotPanic(blk, msg...)
}

func Equal[T any](tb testing.TB, v, oth T, msg ...any) {
	tb.Helper()
	Must(tb).Equal(v, oth, msg...)
}

func NotEqual[T any](tb testing.TB, v, oth T, msg ...any) {
	tb.Helper()
	Must(tb).NotEqual(v, oth, msg...)
}

func Contain(tb testing.TB, haystack, needle any, msg ...any) {
	tb.Helper()
	Must(tb).Contain(haystack, needle, msg...)
}

func NotContain(tb testing.TB, haystack, v any, msg ...any) {
	tb.Helper()
	Must(tb).NotContain(haystack, v, msg...)
}

func ContainExactly[T any /* Map or Slice */](tb testing.TB, v, oth T, msg ...any) {
	tb.Helper()
	Must(tb).ContainExactly(v, oth, msg...)
}

func Sub[T any](tb testing.TB, haystack, needle []T, msg ...any) {
	tb.Helper()
	Must(tb).Sub(haystack, needle, msg...)
}

func ErrorIs(tb testing.TB, err, oth error, msg ...any) {
	tb.Helper()
	Must(tb).ErrorIs(err, oth, msg...)
}

func Error(tb testing.TB, err error, msg ...any) {
	tb.Helper()
	Must(tb).Error(err, msg...)
}

func NoError(tb testing.TB, err error, msg ...any) {
	tb.Helper()
	Must(tb).NoError(err, msg...)
}

func Read[T string | []byte](tb testing.TB, v T, r io.Reader, msg ...any) {
	tb.Helper()
	Must(tb).Read(v, r, msg...)
}

func ReadAll(tb testing.TB, r io.Reader, msg ...any) []byte {
	tb.Helper()
	return Must(tb).ReadAll(r, msg...)
}

func Within(tb testing.TB, timeout time.Duration, blk func(context.Context), msg ...any) {
	tb.Helper()
	Must(tb).Within(timeout, blk, msg...)
}

func NotWithin(tb testing.TB, timeout time.Duration, blk func(context.Context), msg ...any) {
	tb.Helper()
	Must(tb).NotWithin(timeout, blk, msg...)
}

func Match[T string | []byte](tb testing.TB, v T, expr string, msg ...any) {
	tb.Helper()
	Must(tb).Match(string(v), expr, msg...)
}

func NotMatch[T string | []byte](tb testing.TB, v T, expr string, msg ...any) {
	tb.Helper()
	Must(tb).NotMatch(string(v), expr, msg...)
}
