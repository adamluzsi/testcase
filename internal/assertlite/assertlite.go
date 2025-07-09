package assertlite

import (
	"reflect"
	"strings"
	"testing"
)

func True(tb testing.TB, ok bool, msg ...any) {
	tb.Helper()

	if !ok {
		tb.Fatal(msg...)
	}
}

func False(tb testing.TB, nok bool, msg ...any) {
	tb.Helper()

	True(tb, !nok, msg...)
}

func Equal[T any](tb testing.TB, x, y T, msg ...any) {
	tb.Helper()

	True(tb, reflect.DeepEqual(x, y), msg...)
}

func Contains(tb testing.TB, haystack, needle string) {
	tb.Helper()

	if !strings.Contains(haystack, needle) {
		tb.Fatalf("\nhaystack: %#v\nneedle: %#v\n", haystack, needle)
	}
}

func NotContains(tb testing.TB, haystack, needle string) {
	tb.Helper()

	if strings.Contains(haystack, needle) {
		tb.Fatalf("\nShould not contain!\nhaystack: %#v\nneedle: %#v\n", haystack, needle)
	}
}
