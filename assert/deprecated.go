package assert

import (
	"fmt"
	"io"
	"reflect"
	"testing"

	"go.llib.dev/testcase/internal/fmterror"
)

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

// Read
//
// Deprecated: will be removed, use Asserter#ReadAll and Asserter#Equal together instead
func (a Asserter) Read(v any /* string | []byte */, r io.Reader, msg ...Message) {
	const FnMethod = "Read"
	a.TB.Helper()
	if r == nil {
		a.failWith(fmterror.Message{
			Name:    FnMethod,
			Cause:   "io.Reader is nil",
			Message: toMsg(msg),
		})
		return
	}
	content, err := io.ReadAll(r)
	if err != nil {
		a.failWith(fmterror.Message{
			Name:    FnMethod,
			Cause:   "Error occurred during io.Reader.Read",
			Message: toMsg(msg),
			Values: []fmterror.Value{
				{Label: "value", Value: err},
				{Label: "error", Value: err.Error()},
			},
		})
		return
	}
	var val, got any
	switch v := v.(type) {
	case string:
		val = v
		got = string(content)
	case []byte:
		val = v
		got = content
	default:
		a.TB.Fatalf("only string and []byte is supported, not %T", v)
		return
	}
	if a.eq(val, got) {
		pass(a.TB)
		return
	}
	a.failWith(fmterror.Message{
		Name:    FnMethod,
		Cause:   "Read output is not as expected.",
		Message: toMsg(msg),
		Values: []fmterror.Value{
			{Label: "expected value", Value: val},
			{Label: "io.Reader content", Value: got},
		},
	})
}

var oneOfSupportedKinds = map[reflect.Kind]struct{}{
	reflect.Slice: {},
	reflect.Array: {},
}

// OneOf evaluates whether at least one element within the given values meets the conditions set in the assertion block.
//
// Deprecated: use assert.OneOf[T] instead.
func (a Asserter) OneOf(values any, blk /* func( */ any, msg ...Message) {
	tb := a.TB
	tb.Helper()

	vs := reflect.ValueOf(values)
	_, ok := oneOfSupportedKinds[vs.Kind()]
	Must(tb).True(ok, Message(fmt.Sprintf("unexpected list value type: %s", vs.Kind().String())))

	var fnErrMsg = Message(fmt.Sprintf("invalid function signature\n\nExpected:\nfunc(it assert.It, v %s)", vs.Type().Elem()))
	fn := reflect.ValueOf(blk)
	Must(tb).Equal(fn.Kind(), reflect.Func, "blk argument must be a function")
	Must(tb).Equal(fn.Type().NumIn(), 2, fnErrMsg)
	Must(tb).Equal(fn.Type().In(0), reflect.TypeOf((*testing.TB)(nil)).Elem(), fnErrMsg)
	Must(tb).Equal(fn.Type().In(1), vs.Type().Elem(), fnErrMsg)

	a.AnyOf(func(a *A) {
		tb.Helper()
		a.Name = "OneOf"
		a.Cause = "None of the element matched the expectations"

		for i := 0; i < vs.Len(); i++ {
			e := vs.Index(i)
			a.Case(func(it testing.TB) {
				fn.Call([]reflect.Value{reflect.ValueOf(it), e})
			})
			if a.OK() {
				break
			}
		}
	}, msg...)
}
