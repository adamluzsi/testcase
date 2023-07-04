package assert

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/internal/reflects"
	"github.com/adamluzsi/testcase/sandbox"

	"github.com/adamluzsi/testcase/internal/fmterror"
)

func Should(tb testing.TB) Asserter {
	return Asserter{
		TB:   tb,
		Fail: tb.Fail,
	}
}

func Must(tb testing.TB) Asserter {
	return Asserter{
		TB:   tb,
		Fail: tb.FailNow,
	}
}

type Asserter struct {
	TB   testing.TB
	Fail func()
}

func (a Asserter) fn(s any) {
	a.TB.Helper()
	a.TB.Log(s)
	a.Fail()
}

func (a Asserter) try(blk func(a Asserter)) (ok bool) {
	a.TB.Helper()
	dtb := &doubles.TB{}
	blk(Should(dtb))
	return !dtb.IsFailed
}

func (a Asserter) True(v bool, msg ...any) {
	a.TB.Helper()
	if v {
		return
	}
	a.fn(fmterror.Message{
		Method:  "True",
		Cause:   `"true" was expected.`,
		Message: msg,
		Values: []fmterror.Value{
			{
				Label: "value",
				Value: v,
			},
		},
	}.String())
}

func (a Asserter) False(v bool, msg ...any) {
	a.TB.Helper()
	if !a.try(func(a Asserter) { a.True(v) }) {
		return
	}
	a.fn(fmterror.Message{
		Method:  "False",
		Cause:   `"false" was expected.`,
		Message: msg,
		Values: []fmterror.Value{
			{
				Label: "value",
				Value: v,
			},
		},
	}.String())
}

func (a Asserter) Nil(v any, msg ...any) {
	a.TB.Helper()
	if v == nil {
		return
	}
	if reflects.IsNil(v) {
		return
	}
	a.fn(fmterror.Message{
		Method:  "Nil",
		Cause:   "Not nil value received",
		Message: msg,
		Values: []fmterror.Value{
			{
				Label: "value",
				Value: v,
			},
		},
	})
}

func (a Asserter) NotNil(v any, msg ...any) {
	a.TB.Helper()
	if !reflects.IsNil(v) {
		return
	}
	a.fn(fmterror.Message{
		Method:  "NotNil",
		Cause:   "Nil value received",
		Message: msg,
	})
}

func (a Asserter) Panic(blk func(), msg ...any) any {
	a.TB.Helper()
	if ro := sandbox.Run(blk); !ro.OK {
		return ro.PanicValue
	}
	a.fn(fmterror.Message{
		Method:  "Panics",
		Cause:   "Expected to panic or die.",
		Message: msg,
	})
	return nil
}

func (a Asserter) NotPanic(blk func(), msg ...any) {
	a.TB.Helper()
	out := sandbox.Run(blk)
	if out.OK {
		return
	}
	a.fn(fmterror.Message{
		Method:  "Panics",
		Cause:   "Expected to panic or die.",
		Message: msg,
		Values: []fmterror.Value{
			{
				Label: "panic:",
				Value: out.PanicValue,
			},
		},
	})
}

// Equal allows you to match if two entity is equal.
//
// if entities are implementing IsEqual/Equal function, then it will be used to check equality between each other.
//   - value.IsEqual(oth T) bool
//   - value.IsEqual(oth T) (bool, error)
//   - value.Equal(oth T) bool
//   - value.Equal(oth T) (bool, error)
func (a Asserter) Equal(v, oth any, msg ...any) {
	a.TB.Helper()
	const method = "Equal"

	if a.checkTypeEquality(method, v, oth, msg) {
		return
	}

	if a.eq(v, oth) {
		return
	}

	a.TB.Log(fmterror.Message{
		Method:  method,
		Message: msg,
	}.String())
	a.TB.Logf("\n\n%s", DiffFunc(v, oth))
	a.Fail()
}

func (a Asserter) NotEqual(v, oth any, msg ...any) {
	a.TB.Helper()
	const method = "NotEqual"

	if a.checkTypeEquality(method, v, oth, msg) {
		return
	}

	if !a.try(func(a Asserter) { a.Equal(v, oth) }) {
		return
	}

	a.fn(fmterror.Message{
		Method:  method,
		Cause:   "Values are equal.",
		Message: msg,
		Values: []fmterror.Value{
			{
				Label: "value",
				Value: v,
			},
			{
				Label: "other",
				Value: oth,
			},
		},
	}.String())
}

func (a Asserter) checkTypeEquality(method string, v any, oth any, msg []any) (failed bool) {
	a.TB.Helper()
	var (
		vType   = reflect.TypeOf(v)
		othType = reflect.TypeOf(oth)
	)
	if vType == nil || othType == nil {
		return false
	}
	if vType == othType {
		return false
	}
	toRawString := func(rt reflect.Type) fmterror.Raw {
		if rt == nil {
			return "<nil>"
		}
		return fmterror.Raw(rt.String())
	}
	a.TB.Log(fmterror.Message{
		Method:  method,
		Cause:   "incorrect types",
		Message: msg,
		Values: []fmterror.Value{
			{
				Label: "type",
				Value: toRawString(vType),
			},
			{
				Label: "other value's type",
				Value: toRawString(othType),
			},
		},
	}.String())
	a.Fail()
	return true
}

func (a Asserter) eq(exp, act any) bool {
	return eq(a.TB, exp, act)
}

func (a Asserter) Contain(haystack, needle any, msg ...any) {
	a.TB.Helper()
	rSrc := reflect.ValueOf(haystack)
	rHas := reflect.ValueOf(needle)
	if !rSrc.IsValid() {
		a.fn(fmterror.Message{
			Method: "Contain",
			Cause:  "invalid source value",
			Values: []fmterror.Value{
				{Label: "value", Value: haystack},
			},
		}.String())
		return
	}
	if !rHas.IsValid() {
		a.fn(fmterror.Message{
			Method: "Contain",
			Cause:  `invalid "has" value`,
			Values: []fmterror.Value{{Label: "value", Value: needle}},
		}.String())
		return
	}

	switch {
	case rSrc.Kind() == reflect.String && rHas.Kind() == reflect.String:
		a.stringContainsSub(rSrc, rHas, msg)

	case rSrc.Kind() == reflect.Slice && rSrc.Type().Elem() == rHas.Type():
		a.sliceContainsValue(rSrc, rHas, msg)

	case rSrc.Kind() == reflect.Slice && rSrc.Type().Elem().Kind() == reflect.Interface && rHas.Type().Implements(rSrc.Type().Elem()):
		a.sliceContainsValue(rSrc, rHas, msg)

	case rSrc.Kind() == reflect.Slice && rSrc.Type() == rHas.Type():
		a.sliceContainsSubSlice(rSrc, rHas, msg)

	case rSrc.Kind() == reflect.Map && rSrc.Type() == rHas.Type():
		a.mapContainsSubMap(rSrc, rHas, msg)

	default:
		panic(fmterror.Message{
			Method: "Contain",
			Cause:  "Unimplemented scenario or type mismatch.",
			Values: []fmterror.Value{
				{
					Label: "source-type",
					Value: fmt.Sprintf("%T", haystack),
				},
				{
					Label: "value-type",
					Value: fmt.Sprintf("%T", needle),
				},
			},
		}.String())
	}
}

func (a Asserter) failContains(src, sub any, msg ...any) {
	a.TB.Helper()

	a.fn(fmterror.Message{
		Method:  "Contain",
		Cause:   "Source doesn't contains expected value(s).",
		Message: msg,
		Values: []fmterror.Value{
			{
				Label: "source",
				Value: src,
			},
			{
				Label: "sub",
				Value: sub,
			},
		},
	}.String())
}

func (a Asserter) sliceContainsValue(slice, value reflect.Value, msg []any) {
	a.TB.Helper()
	var found bool
	for i := 0; i < slice.Len(); i++ {
		if a.eq(slice.Index(i).Interface(), value.Interface()) {
			found = true
			break
		}
	}
	if found {
		return
	}
	a.fn(fmterror.Message{
		Method:  "Contain",
		Cause:   "Couldn't find the expected value in the source slice",
		Message: msg,
		Values: []fmterror.Value{
			{
				Label: "source",
				Value: slice.Interface(),
			},
			{
				Label: "value",
				Value: value.Interface(),
			},
		},
	})
}

func (a Asserter) sliceContainsSubSlice(haystack, needle reflect.Value, msg []any) {
	a.TB.Helper()

	if haystack.Len() < needle.Len() {
		a.fn(fmterror.Message{
			Method:  "Contain",
			Cause:   "Haystack slice is smaller than needle slice.",
			Message: msg,
			Values: []fmterror.Value{
				{
					Label: "haystack slice len",
					Value: haystack.Len(),
				},
				{
					Label: "needle slice len",
					Value: needle.Len(),
				},
			},
		}.String())
		return
	}

	for i := 0; i < needle.Len(); i++ {
		needleElem := needle.Index(i)
		var found bool

	searchingHaystack:
		for j := 0; j < haystack.Len(); j++ {
			haystackElem := haystack.Index(j)

			if a.eq(haystackElem.Interface(), needleElem.Interface()) {
				found = true
				break searchingHaystack
			}
		}
		if !found {
			a.fn(fmterror.Message{
				Method:  "Contain",
				Cause:   "Haystack slice doesn't contains expected value(s) of needle slice.",
				Message: msg,
				Values: []fmterror.Value{
					{
						Label: "haystack slice",
						Value: haystack.Interface(),
					},
					{
						Label: "needle slice",
						Value: needle.Interface(),
					},
					{
						Label: "missing element",
						Value: needleElem.Interface(),
					},
				},
			}.String())
		}
	}
}

func (a Asserter) Sub(slice, sub any, msg ...any) {
	a.TB.Helper()

	sliceRV := reflect.ValueOf(slice)
	subRV := reflect.ValueOf(sub)

	switch sliceRV.Kind() {
	case reflect.Slice:
	default:
		a.TB.Fatalf("unsuported argument type for .Sub: %T", slice)
		return
	}

	if sliceRV.Type() != subRV.Type() {
		a.TB.Fatalf("argument type mismatch for .Sub: %T / %T", slice, sub)
		return
	}

	failWithNotEqual := func(missingElement any) {
		values := []fmterror.Value{
			{
				Label: "source",
				Value: slice,
			},
			{
				Label: "subset",
				Value: sub,
			},
		}
		if missingElement != nil {
			values = append(values, fmterror.Value{
				Label: "missing element",
				Value: missingElement,
			})
		}
		a.fn(fmterror.Message{
			Method:  "Subset",
			Cause:   "Slice doesn't contain the expected subset.",
			Message: msg,
			Values:  values,
		}.String())
	}

	if sliceRV.Len() < subRV.Len() {
		a.fn(fmterror.Message{
			Method:  "Contain",
			Cause:   "Source slice is smaller than sub slice.",
			Message: msg,
			Values: []fmterror.Value{
				{
					Label: "source",
					Value: sliceRV.Interface(),
				},
				{
					Label: "sub",
					Value: subRV.Interface(),
				},
			},
		}.String())
		return
	}

	var (
		offset int
		found  bool
	)
searching:
	for i := 0; i < sliceRV.Len(); i++ {
		for j := 0; j < subRV.Len(); j++ {
			if a.eq(sliceRV.Index(i).Interface(), subRV.Index(j).Interface()) {
				offset = i
				found = true
				break searching
			}
		}
	}

	if !found {
		failWithNotEqual(nil)
		return
	}

	for i := 0; i < subRV.Len(); i++ {
		expected := sliceRV.Index(i + offset).Interface()
		actual := subRV.Index(i).Interface()

		if !a.eq(expected, actual) {
			failWithNotEqual(actual)
			return
		}
	}
}

// Match will match an expression against a given value.
// Match will fail for both receiving an invalid expression
// or having the value not matched by the expression.
// If the expression is invalid, test will fail early, regardless if Should or Must was used.
func (a Asserter) Match(v, expr string, msg ...any) {
	a.TB.Helper()
	if a.toRegexp(expr).MatchString(v) {
		return
	}
	a.fn(fmterror.Message{
		Method:  "Match",
		Cause:   "failed to match the expected expression",
		Message: msg,
		Values: []fmterror.Value{
			{Label: "value", Value: v},
			{Label: "expression", Value: expr},
		},
	})
}

// NotMatch will check if an expression is not matching a given value.
// NotMatch will fail the test early for receiving an invalid expression.
func (a Asserter) NotMatch(v, expr string, msg ...any) {
	a.TB.Helper()
	if !a.toRegexp(expr).MatchString(v) {
		return
	}
	a.fn(fmterror.Message{
		Method:  "NotMatch",
		Cause:   "value is matching the expression",
		Message: msg,
		Values: []fmterror.Value{
			{Label: "value", Value: v},
			{Label: "expression", Value: expr},
		},
	})
}

func (a Asserter) toRegexp(expr string) *regexp.Regexp {
	a.TB.Helper()
	rgx, err := regexp.Compile(expr)
	if err != nil {
		a.TB.Log(fmterror.Message{
			Method: "NotMatch",
			Cause:  "invalid expression given",
			Values: []fmterror.Value{
				{Label: "expression", Value: expr},
				{Label: "regexp compile error", Value: err},
			},
		})
		a.TB.FailNow()
	}
	return rgx
}

func (a Asserter) mapContainsSubMap(src reflect.Value, has reflect.Value, msg []any) {
	a.TB.Helper()
	for _, key := range has.MapKeys() {
		srcValue := src.MapIndex(key)
		if !srcValue.IsValid() {
			a.fn(fmterror.Message{
				Method:  "Contain",
				Cause:   "Source doesn't contains the other map.",
				Message: msg,
				Values: []fmterror.Value{
					{
						Label: "source",
						Value: src.Interface(),
					},
					{
						Label: "key",
						Value: key.Interface(),
					},
				},
			})
			return
		}
		if !a.eq(srcValue.Interface(), has.MapIndex(key).Interface()) {
			a.fn(fmterror.Message{
				Method:  "Contain",
				Cause:   "Source has the key but with different value.",
				Message: msg,
				Values: []fmterror.Value{
					{
						Label: "source",
						Value: src.Interface(),
					},
					{
						Label: "key",
						Value: key.Interface(),
					},
				},
			})
			return
		}
	}
}

func (a Asserter) stringContainsSub(src reflect.Value, has reflect.Value, msg []any) {
	a.TB.Helper()
	if strings.Contains(fmt.Sprint(src.Interface()), fmt.Sprint(has.Interface())) {
		return
	}
	a.fn(fmterror.Message{
		Method:  "Contain",
		Cause:   "String doesn't include sub string.",
		Message: msg,
		Values: []fmterror.Value{
			{
				Label: "string",
				Value: src.Interface(),
			},
			{
				Label: "substr",
				Value: has.Interface(),
			},
		},
	})
}

func (a Asserter) NotContain(haystack, v any, msg ...any) {
	a.TB.Helper()
	if !a.try(func(a Asserter) { a.Contain(haystack, v) }) {
		return
	}
	a.fn(fmterror.Message{
		Method:  "NotContain",
		Cause:   "Source contains the received value",
		Message: msg,
		Values: []fmterror.Value{
			{
				Label: "haystack",
				Value: haystack,
			},
			{
				Label: "other",
				Value: v,
			},
		},
	})
}

func (a Asserter) ContainExactly(v, oth any /* slice | map */, msg ...any) {
	a.TB.Helper()

	rv := reflect.ValueOf(v)
	roth := reflect.ValueOf(oth)

	if !rv.IsValid() {
		panic(fmterror.Message{
			Method: "ContainExactly",
			Cause:  "invalid expected value",
			Values: []fmterror.Value{
				{
					Label: "value",
					Value: v,
				},
			},
		}.String())
	}
	if !roth.IsValid() {
		panic(fmterror.Message{
			Method: "ContainExactly",
			Cause:  `invalid actual value`,
			Values: []fmterror.Value{
				{
					Label: "value",
					Value: oth,
				},
			},
		}.String())
	}

	switch {
	case rv.Kind() == reflect.Slice && rv.Type() == roth.Type():
		a.containExactlySlice(rv, roth, msg)

	case rv.Kind() == reflect.Map && rv.Type() == roth.Type():
		a.containExactlyMap(rv, roth, msg)

	default:
		panic(fmterror.Message{
			Method: "ContainExactly",
			Cause:  "invalid type, slice or map was expected",
			Values: []fmterror.Value{
				{
					Label: "type of the value",
					Value: fmterror.Raw(fmt.Sprintf("%T", v)),
				},
				{
					Label: "kind of the value",
					Value: fmterror.Raw(rv.Kind().String()),
				},
				{
					Label: "type of the other value",
					Value: fmterror.Raw(fmt.Sprintf("%T", oth)),
				},
				{
					Label: "kind of the other value",
					Value: fmterror.Raw(roth.Kind().String()),
				},
			},
		}.String())
	}
}

func (a Asserter) containExactlyMap(exp reflect.Value, act reflect.Value, msg []any) {
	a.TB.Helper()

	if a.eq(exp.Interface(), act.Interface()) {
		return
	}
	a.fn(fmterror.Message{
		Method:  "ContainExactly",
		Cause:   "SubMap content doesn't exactly match with expectations.",
		Message: msg,
		Values: []fmterror.Value{
			{Label: "expected", Value: exp.Interface()},
			{Label: "actual", Value: act.Interface()},
		},
	})
}

func (a Asserter) containExactlySlice(exp reflect.Value, act reflect.Value, msg []any) {
	a.TB.Helper()

	if exp.Len() != act.Len() {
		a.fn(fmterror.Message{
			Method:  "ContainExactly",
			Cause:   "Element count doesn't match",
			Message: msg,
			Values: []fmterror.Value{
				{
					Label: "expected:",
					Value: exp.Len(),
				},
				{
					Label: "actual:",
					Value: act.Len(),
				},
			},
		})
	}

	for i := 0; i < exp.Len(); i++ {
		expectedValue := exp.Index(i).Interface()

		var found bool
	search:
		for j := 0; j < act.Len(); j++ {
			if a.eq(expectedValue, act.Index(j).Interface()) {
				found = true
				break search
			}
		}
		if !found {
			a.fn(fmterror.Message{
				Method:  "ContainExactly",
				Cause:   fmt.Sprintf("Element not found at index %d", i),
				Message: msg,
				Values: []fmterror.Value{
					{
						Label: "actual:",
						Value: act.Interface(),
					},
					{
						Label: "value",
						Value: expectedValue,
					},
				},
			})
		}
	}
}

func (a Asserter) AnyOf(blk func(a *AnyOf), msg ...any) {
	a.TB.Helper()
	anyOf := &AnyOf{TB: a.TB, Fail: a.Fail}
	defer anyOf.Finish(msg...)
	blk(anyOf)
}

var timeType = reflect.TypeOf(time.Time{})

func (a Asserter) isEmpty(v any) bool {
	a.TB.Helper()
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Map, reflect.Slice:
		return rv.Len() == 0
	case reflect.Array:
		zero := reflect.New(rv.Type()).Elem().Interface()
		return a.eq(zero, v)
	case reflect.Ptr:
		if rv.IsNil() {
			return true
		}
		return a.isEmpty(rv.Elem().Interface())
	default:
		if tm, ok := v.(time.Time); ok {
			return tm.IsZero()
		}
		return a.eq(reflect.Zero(rv.Type()).Interface(), v)
	}
}

// Empty gets whether the specified value is considered empty.
func (a Asserter) Empty(v any, msg ...any) {
	a.TB.Helper()
	if a.isEmpty(v) {
		return
	}
	a.fn(fmterror.Message{
		Method:  "Empty",
		Cause:   "Value was expected to be empty.",
		Message: msg,
		Values: []fmterror.Value{
			{Label: "value", Value: v},
		},
	})
}

// NotEmpty gets whether the specified value is considered empty.
func (a Asserter) NotEmpty(v any, msg ...any) {
	a.TB.Helper()
	if !a.isEmpty(v) {
		return
	}
	a.fn(fmterror.Message{
		Method:  "NotEmpty",
		Cause:   "Value was expected to be not empty.",
		Message: msg,
		Values: []fmterror.Value{
			{Label: "value", Value: v},
		},
	})
}

// ErrorIs allow you to assert an error value by an expectation.
// ErrorIs allow asserting an error regardless if it's wrapped or not.
// Suppose the implementation of the test subject later changes by wrap errors to add more context to the return error.
func (a Asserter) ErrorIs(err, oth error, msg ...any) {
	a.TB.Helper()
	if a.errorIs(err, oth) || a.errorIs(oth, err) {
		return
	}
	a.fn(fmterror.Message{
		Method:  "ErrorIs",
		Cause:   "error value is not what was expected",
		Message: msg,
		Values: []fmterror.Value{
			{Label: "err", Value: err},
			{Label: "oth", Value: oth},
		},
	})
}

func (a Asserter) errorIs(err, oth error) bool {
	a.TB.Helper()
	if err == nil && oth == nil {
		return true
	}
	if errors.Is(err, oth) {
		return true
	}
	if a.eq(oth, err) {
		return true
	}
	if oth != nil {
		if ptr := reflect.New(reflect.TypeOf(oth)); errors.As(err, ptr.Interface()) {
			return a.eq(oth, ptr.Elem().Interface())
		}
	}
	return false
}

func (a Asserter) Error(err error, msg ...any) {
	a.TB.Helper()
	if err != nil {
		return
	}
	a.fn(fmterror.Message{
		Method:  "Error",
		Cause:   "Expected an error, but got nil.",
		Message: msg,
		Values: []fmterror.Value{
			{Label: "value", Value: err},
		},
	})
}

func (a Asserter) NoError(err error, msg ...any) {
	a.TB.Helper()
	if err == nil {
		return
	}
	a.fn(fmterror.Message{
		Method:  "NoError",
		Cause:   "Non-nil error value is received.",
		Message: msg,
		Values: []fmterror.Value{
			{Label: "value", Value: err},
			{Label: "error", Value: err.Error()},
		},
	})
}

func (a Asserter) Read(v any /* string | []byte */, r io.Reader, msg ...any) {
	const FnMethod = "Read"
	a.TB.Helper()
	if r == nil {
		a.fn(fmterror.Message{
			Method:  FnMethod,
			Cause:   "io.Reader is nil",
			Message: msg,
		})
		return
	}
	content, err := io.ReadAll(r)
	if err != nil {
		a.fn(fmterror.Message{
			Method:  FnMethod,
			Cause:   "Error occurred during io.Reader.Read",
			Message: msg,
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
		return
	}
	a.fn(fmterror.Message{
		Method:  FnMethod,
		Cause:   "Read output is not as expected.",
		Message: msg,
		Values: []fmterror.Value{
			{Label: "expected value", Value: val},
			{Label: "io.Reader content", Value: got},
		},
	})
}

func (a Asserter) ReadAll(r io.Reader, msg ...any) []byte {
	a.TB.Helper()
	const FnMethod = "ReadAll"
	if r == nil {
		a.fn(fmterror.Message{
			Method:  FnMethod,
			Cause:   "io.Reader is nil",
			Message: msg,
		})
		return nil
	}
	bs, err := io.ReadAll(r)
	if err != nil {
		a.fn(fmterror.Message{
			Method:  FnMethod,
			Cause:   "Error occurred during io.ReadAll",
			Message: msg,
			Values: []fmterror.Value{
				{Label: "value", Value: err},
				{Label: "error", Value: err.Error()},
			},
		})
		return nil
	}
	return bs
}

func (a Asserter) Within(timeout time.Duration, blk func(context.Context), msg ...any) {
	a.TB.Helper()
	if !a.within(timeout, blk) {
		a.fn(fmterror.Message{
			Method:  "Within",
			Cause:   "Expected to finish within the timeout duration.",
			Message: msg,
			Values: []fmterror.Value{
				{
					Label: "timeout",
					Value: timeout,
				},
			},
		}.String())
	}
}

func (a Asserter) NotWithin(timeout time.Duration, blk func(context.Context), msg ...any) {
	a.TB.Helper()
	if a.within(timeout, blk) {
		a.fn(fmterror.Message{
			Method:  "NotWithin",
			Cause:   `Expected to not finish within the timeout duration.`,
			Message: msg,
			Values: []fmterror.Value{
				{
					Label: "timeout",
					Value: timeout,
				},
			},
		}.String())
	}
}

func (a Asserter) within(timeout time.Duration, blk func(context.Context)) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var done, isFailNow uint32
	go func() {
		ro := sandbox.Run(func() {
			blk(ctx)
			atomic.AddUint32(&done, 1)
		})
		if !ro.OK {
			atomic.AddUint32(&isFailNow, 1)
		}
	}()
	Waiter{Timeout: timeout}.While(func() bool {
		return atomic.LoadUint32(&done) == 0 && atomic.LoadUint32(&isFailNow) == 0
	})
	if atomic.LoadUint32(&isFailNow) != 0 {
		a.TB.FailNow()
	}
	return atomic.LoadUint32(&done) == 1
}
