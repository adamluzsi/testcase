package assert

import (
	"io"
	"testing"
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

func Equal[T any](tb testing.TB, expected, actually T, msg ...any) {
	tb.Helper()
	Must(tb).Equal(expected, actually, msg...)
}

func NotEqual[T any](tb testing.TB, expected, actually T, msg ...any) {
	tb.Helper()
	Must(tb).NotEqual(expected, actually, msg...)
}

func Contain(tb testing.TB, haystack, needle any, msg ...any) {
	tb.Helper()
	Must(tb).Contain(haystack, needle, msg...)
}

func NotContain(tb testing.TB, haystack, v any, msg ...any) {
	tb.Helper()
	Must(tb).NotContain(haystack, v, msg...)
}

func ContainExactly[T any](tb testing.TB, expected, actual T, msg ...any) {
	tb.Helper()
	Must(tb).ContainExactly(expected, actual, msg...)
}

func Sub[T any](tb testing.TB, haystack, needle []T, msg ...any) {
	tb.Helper()
	Must(tb).Sub(haystack, needle, msg...)
}

func ErrorIs(tb testing.TB, expected, actual error, msg ...any) {
	tb.Helper()
	Must(tb).ErrorIs(expected, actual, msg...)
}

func Error(tb testing.TB, err error, msg ...any) {
	tb.Helper()
	Must(tb).Error(err, msg...)
}

func NoError(tb testing.TB, err error, msg ...any) {
	tb.Helper()
	Must(tb).NoError(err, msg...)
}

func Read[T string | []byte](tb testing.TB, expected T, r io.Reader, msg ...any) {
	tb.Helper()
	Must(tb).Read(expected, r, msg...)
}

func ReadAll(tb testing.TB, r io.Reader, msg ...any) []byte {
	tb.Helper()
	return Must(tb).ReadAll(r, msg...)
}
