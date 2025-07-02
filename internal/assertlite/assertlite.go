package assertlite

import (
	"reflect"
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
