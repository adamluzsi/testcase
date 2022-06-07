package assert

import (
	"testing"
)

func True(tb testing.TB, v bool, msg ...any) {
	Must(tb).True(v, msg...)
}

func False(tb testing.TB, v bool, msg ...any) {
	Must(tb).False(v, msg...)
}

func Nil(tb testing.TB, v any, msg ...any) {
	Must(tb).Nil(v, msg...)
}

func NotNil(tb testing.TB, v any, msg ...any) {
	Must(tb).NotNil(v, msg...)
}

func Empty(tb testing.TB, v any, msg ...any) {
	Must(tb).Empty(v, msg...)
}

func NotEmpty(tb testing.TB, v any, msg ...any) {
	Must(tb).NotEmpty(v, msg...)
}

func Panic(tb testing.TB, blk func(), msg ...any) any {
	return Must(tb).Panic(blk, msg...)
}

func NotPanic(tb testing.TB, blk func(), msg ...any) {
	Must(tb).NotPanic(blk, msg...)
}

func Equal[T any](tb testing.TB, expected, actually T, msg ...any) {
	Must(tb).Equal(expected, actually, msg...)
}

func NotEqual[T any](tb testing.TB, expected, actually T, msg ...any) {
	Must(tb).NotEqual(expected, actually, msg...)
}

func Contain(tb testing.TB, haystack, needle any, msg ...any) {
	Must(tb).Contain(haystack, needle, msg...)
}

func NotContain(tb testing.TB, haystack, v any, msg ...any) {
	Must(tb).NotContain(haystack, v, msg...)
}

func ContainExactly[T any](tb testing.TB, expected, actual T, msg ...any) {
	Must(tb).ContainExactly(expected, actual, msg...)
}

func ErrorIs(tb testing.TB, expected, actual error, msg ...any) {
	Must(tb).ErrorIs(expected, actual, msg...)
}
