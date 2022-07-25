package assert

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/internal/reflects"
	"github.com/adamluzsi/testcase/pp"
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
	if !a.try(func(a Asserter) { a.Nil(v) }) {
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
// if entities are implementing IsEqual function, then it will be used to check equality between each other.
//   - IsEqual(oth T) bool
//   - IsEqual(oth T) (bool, error)
//
func (a Asserter) Equal(expected, actually any, msg ...any) {
	a.TB.Helper()
	if a.eq(expected, actually) {
		return
	}

	a.TB.Log(fmterror.Message{
		Method:  "Equal",
		Message: msg,
		Values: []fmterror.Value{
			{
				Label: "expected",
				Value: expected,
			},
			{
				Label: "actual",
				Value: actually,
			},
		},
	}.String())
	a.TB.Logf("\n\n%s", pp.Diff(expected, actually))
	a.Fail()
}

func (a Asserter) NotEqual(v, oth any, msg ...any) {
	a.TB.Helper()
	if !a.try(func(a Asserter) { a.Equal(v, oth) }) {
		return
	}
	a.fn(fmterror.Message{
		Method:  "NotEqual",
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

func (a Asserter) eq(exp, act any) bool {
	if isEqual, ok := a.tryIsEqual(exp, act); ok {
		return isEqual
	}

	return reflect.DeepEqual(exp, act)
}

func (a Asserter) tryIsEqual(exp, act any) (isEqual bool, ok bool) {
	defer func() { recover() }()
	expRV := reflect.ValueOf(exp)
	actRV := reflect.ValueOf(act)

	if expRV.Type() != actRV.Type() {
		return false, false
	}

	method := expRV.MethodByName("IsEqual")
	methodType := method.Type()

	if methodType.NumIn() != 1 {
		return false, false
	}
	if numOut := methodType.NumOut(); !(numOut == 1 || numOut == 2) {
		return false, false
	}
	if methodType.In(0) != actRV.Type() {
		return false, false
	}

	res := method.Call([]reflect.Value{actRV})

	switch {
	case methodType.NumOut() == 1: // IsEqual(T) (bool)
		return res[0].Bool(), true

	case methodType.NumOut() == 2: // IsEqual(T) (bool, error)
		Must(a.TB).Nil(res[1].Interface())
		return res[0].Bool(), true

	default:
		return false, false
	}
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

func (a Asserter) sliceContainsSubSlice(slice, sub reflect.Value, msg []any) {
	a.TB.Helper()

	failWithNotEqual := func() { a.failContains(slice.Interface(), sub.Interface(), msg...) }

	if slice.Len() < sub.Len() {
		a.fn(fmterror.Message{
			Method:  "Contain",
			Cause:   "Source slice is smaller than sub slice.",
			Message: msg,
			Values: []fmterror.Value{
				{
					Label: "source",
					Value: slice.Interface(),
				},
				{
					Label: "sub",
					Value: sub.Interface(),
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
	for i := 0; i < slice.Len(); i++ {
		for j := 0; j < sub.Len(); j++ {
			if a.eq(slice.Index(i).Interface(), sub.Index(j).Interface()) {
				offset = i
				found = true
				break searching
			}
		}
	}

	if !found {
		failWithNotEqual()
		return
	}

	for i := 0; i < sub.Len(); i++ {
		expected := slice.Index(i + offset).Interface()
		actual := sub.Index(i).Interface()

		if !a.eq(expected, actual) {
			failWithNotEqual()
			return
		}
	}
}

func (a Asserter) mapContainsSubMap(src reflect.Value, has reflect.Value, msg []any) {
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

func (a Asserter) ContainExactly(expected, actual any, msg ...any) {
	a.TB.Helper()

	exp := reflect.ValueOf(expected)
	act := reflect.ValueOf(actual)

	if !exp.IsValid() {
		panic(fmterror.Message{
			Method: "ContainExactly",
			Cause:  "invalid expected value",
			Values: []fmterror.Value{
				{
					Label: "value",
					Value: expected,
				},
			},
		}.String())
	}
	if !act.IsValid() {
		panic(fmterror.Message{
			Method: "ContainExactly",
			Cause:  `invalid actual value`,
			Values: []fmterror.Value{
				{
					Label: "value",
					Value: actual,
				},
			},
		}.String())
	}

	switch {
	case exp.Kind() == reflect.Slice && exp.Type() == act.Type():
		a.containExactlySlice(exp, act, msg)

	case exp.Kind() == reflect.Map && exp.Type() == act.Type():
		a.containExactlyMap(exp, act, msg)

	default:
		// TODO: maybe use Equal as default approach?
		panic(fmterror.Message{
			Method: "ContainExactly",
			Cause:  "Unimplemented scenario or type mismatch.",
			Values: []fmterror.Value{
				{
					Label: "expected-type",
					Value: fmt.Sprintf("%T", expected),
				},
				{
					Label: "actual-type",
					Value: fmt.Sprintf("%T", actual),
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
					Label: "actual:",
					Value: act.Interface(),
				},
				{
					Label: "value",
					Value: exp.Interface(),
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
	anyOf := &AnyOf{TB: a.TB, Fail: a.TB.Fail}
	defer anyOf.Finish(msg...)
	blk(anyOf)
}

func (a Asserter) isEmpty(v any) bool {
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
	if !a.try(func(a Asserter) { a.Empty(v) }) {
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
func (a Asserter) ErrorIs(expected, actual error, msg ...any) {
	a.TB.Helper()

	if errors.Is(actual, expected) {
		return
	}
	if a.eq(expected, actual) {
		return
	}
	if ErrorEqAs := func(expected, actual error) bool {
		if actual == nil || expected == nil {
			return false
		}
		nErr := reflect.New(reflect.TypeOf(expected))
		return errors.As(actual, nErr.Interface()) &&
			a.eq(expected, nErr.Elem().Interface())
	}; ErrorEqAs(expected, actual) {
		return
	}

	a.fn(fmterror.Message{
		Method:  "ErrorIs",
		Cause:   "The actual error is not what was expected.",
		Message: msg,
		Values: []fmterror.Value{
			{Label: "expected", Value: expected},
			{Label: "actual", Value: actual},
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

func (a Asserter) Read(expected any, r io.Reader, msg ...any) {
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
	actual, err := io.ReadAll(r)
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
	var exp, act any
	switch v := expected.(type) {
	case string:
		exp = v
		act = string(actual)
	case []byte:
		exp = v
		act = actual
	default:
		a.TB.Fatalf("only string and []byte is supported, not %T", v)
		return
	}
	if a.eq(exp, act) {
		return
	}
	a.fn(fmterror.Message{
		Method:  FnMethod,
		Cause:   "Read output is not as expected.",
		Message: msg,
		Values: []fmterror.Value{
			{Label: "expected", Value: exp},
			{Label: "actual", Value: act},
		},
	})
}

func (a Asserter) ReadAll(r io.Reader, msg ...any) []byte {
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
