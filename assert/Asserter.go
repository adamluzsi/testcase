package assert

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/adamluzsi/testcase/internal"

	"github.com/adamluzsi/testcase/internal/fmterror"
)

func Should(tb testing.TB) Asserter {
	return Asserter{
		TB: tb,
		Fn: tb.Error,
	}
}

func Must(tb testing.TB) Asserter {
	return Asserter{
		TB: tb,
		Fn: tb.Fatal,
	}
}

type Asserter struct {
	TB testing.TB
	Fn func(args ...interface{})
}

func (a Asserter) try(blk func(a Asserter)) (ok bool) {
	a.TB.Helper()
	var failed bool
	blk(Asserter{TB: a.TB, Fn: func(args ...interface{}) { failed = true }})
	return !failed
}

func (a Asserter) True(v bool, msg ...interface{}) {
	a.TB.Helper()
	if v {
		return
	}
	a.Fn(fmterror.Message{
		Method: "True",
		Cause:  `"true" was expected.`,
		Values: []fmterror.Value{
			{
				Label: "value",
				Value: v,
			},
		},
		UserMessage: msg,
	}.String())
}

func (a Asserter) False(v bool, msg ...interface{}) {
	a.TB.Helper()
	if !a.try(func(a Asserter) { a.True(v) }) {
		return
	}
	a.Fn(fmterror.Message{
		Method: "False",
		Cause:  `"false" was expected.`,
		Values: []fmterror.Value{
			{
				Label: "value",
				Value: v,
			},
		},
		UserMessage: msg,
	}.String())
}

func (a Asserter) Nil(v interface{}, msg ...interface{}) {
	a.TB.Helper()
	if v == nil {
		return
	}
	if func() (isNil bool) {
		defer func() { _ = recover() }()

		return reflect.ValueOf(v).IsNil()
	}() {
		return
	}
	a.Fn(fmterror.Message{
		Method: "Nil",
		Cause:  "Not nil value received",
		Values: []fmterror.Value{
			{
				Label: "value",
				Value: v,
			},
		},
		UserMessage: msg,
	})
}

func (a Asserter) NotNil(v interface{}, msg ...interface{}) {
	a.TB.Helper()
	if !a.try(func(a Asserter) { a.Nil(v) }) {
		return
	}
	a.Fn(fmterror.Message{
		Method:      "NotNil",
		Cause:       "Nil value received",
		UserMessage: msg,
	})
}

func (a Asserter) hasPanicked(blk func()) (panicValue interface{}, ok bool) {
	a.TB.Helper()
	panicValue, finished := internal.Recover(blk)
	return panicValue, !finished
}

func (a Asserter) Panic(blk func(), msg ...interface{}) (panicValue interface{}) {
	a.TB.Helper()
	panicValue, ok := a.hasPanicked(blk)
	if ok {
		return panicValue
	}
	a.Fn(fmterror.Message{
		Method:      "Panics",
		Cause:       "Expected to panic or die.",
		UserMessage: msg,
	})
	return nil
}

func (a Asserter) NotPanic(blk func(), msg ...interface{}) {
	a.TB.Helper()
	panicValue, ok := a.hasPanicked(blk)
	if !ok {
		return
	}
	a.Fn(fmterror.Message{
		Method: "Panics",
		Cause:  "Expected to panic or die.",
		Values: []fmterror.Value{
			{
				Label: "panic:",
				Value: panicValue,
			},
		},
		UserMessage: msg,
	})
}

type equalable[T any] interface {
	IsEqual(oth T) bool
}

type equalableWithError[T any] interface {
	IsEqual(oth T) (bool, error)
}

func (a Asserter) Equal(expected, actually interface{}, msg ...interface{}) {
	a.TB.Helper()
	if a.eq(expected, actually) {
		return
	}
	a.Fn(fmterror.Message{
		Method: "Equal",
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
		UserMessage: msg,
	}.String())
}

func (a Asserter) NotEqual(v, oth interface{}, msg ...interface{}) {
	a.TB.Helper()
	if !a.try(func(a Asserter) { a.Equal(v, oth) }) {
		return
	}
	a.Fn(fmterror.Message{
		Method: "NotEqual",
		Cause:  "Values are equal.",
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
		UserMessage: msg,
	}.String())
}

func (a Asserter) eq(exp, act interface{}) bool {
	if isEqual, ok := a.tryIsEqual(exp, act); ok {
		return isEqual
	}

	return reflect.DeepEqual(exp, act)
}

func (a Asserter) tryIsEqual(exp, act interface{}) (isEqual bool, ok bool) {
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

func (a Asserter) Contain(src, has interface{}, msg ...interface{}) {
	a.TB.Helper()
	rSrc := reflect.ValueOf(src)
	rHas := reflect.ValueOf(has)
	if !rSrc.IsValid() {
		a.Fn(fmterror.Message{
			Method: "Contains",
			Cause:  "invalid source value",
			Values: []fmterror.Value{
				{Label: "value", Value: src},
			},
		}.String())
		return
	}
	if !rHas.IsValid() {
		a.Fn(fmterror.Message{
			Method: "Contains",
			Cause:  `invalid "has" value`,
			Values: []fmterror.Value{{Label: "value", Value: has}},
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
			Method: "Contains",
			Cause:  "Unimplemented scenario or type mismatch.",
			Values: []fmterror.Value{
				{
					Label: "source-type",
					Value: fmt.Sprintf("%T", src),
				},
				{
					Label: "value-type",
					Value: fmt.Sprintf("%T", has),
				},
			},
		}.String())
	}
}

func (a Asserter) failContains(src, sub interface{}, msg ...interface{}) {
	a.TB.Helper()

	a.Fn(fmterror.Message{
		Method: "Contains",
		Cause:  "Source doesn't contains expected value(s).",
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
		UserMessage: msg,
	}.String())
}

func (a Asserter) sliceContainsValue(slice, value reflect.Value, msg []interface{}) {
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
	a.Fn(fmterror.Message{
		Method: "Contains",
		Cause:  "Couldn't find the expected value in the source slice",
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
		UserMessage: msg,
	})
}

func (a Asserter) sliceContainsSubSlice(slice, sub reflect.Value, msg []interface{}) {
	a.TB.Helper()

	failWithNotEqual := func() { a.failContains(slice.Interface(), sub.Interface(), msg...) }

	//if slice.Kind() != reflect.Slice || sub.Kind() != reflect.Slice {
	//	a.Fn(fmterror.Message{
	//		Method: "Contains",
	//		Cause:  "Invalid slice type(s).",
	//		Values: []fmterror.Value{
	//			{
	//				Label: "source",
	//				Value: slice.Interface(),
	//			},
	//			{
	//				Label: "sub",
	//				Value: sub.Interface(),
	//			},
	//		},
	//		UserMessage: msg,
	//	}.String())
	//	return
	//}
	if slice.Len() < sub.Len() {
		a.Fn(fmterror.Message{
			Method: "Contains",
			Cause:  "Source slice is smaller than sub slice.",
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
			UserMessage: msg,
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

func (a Asserter) mapContainsSubMap(src reflect.Value, has reflect.Value, msg []interface{}) {
	for _, key := range has.MapKeys() {
		srcValue := src.MapIndex(key)
		if !srcValue.IsValid() {
			a.Fn(fmterror.Message{
				Method: "Contains",
				Cause:  "Source doesn't contains the other map.",
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
				UserMessage: msg,
			})
			return
		}
		if !a.eq(srcValue.Interface(), has.MapIndex(key).Interface()) {
			a.Fn(fmterror.Message{
				Method: "Contains",
				Cause:  "Source has the key but with different value.",
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
				UserMessage: msg,
			})
			return
		}
	}
}

func (a Asserter) stringContainsSub(src reflect.Value, has reflect.Value, msg []interface{}) {
	a.TB.Helper()
	if strings.Contains(fmt.Sprint(src.Interface()), fmt.Sprint(has.Interface())) {
		return
	}
	a.Fn(fmterror.Message{
		Method: "Contains",
		Cause:  "String doesn't include sub string.",
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
		UserMessage: msg,
	})
}

func (a Asserter) NotContain(source, oth interface{}, msg ...interface{}) {
	a.TB.Helper()
	if !a.try(func(a Asserter) { a.Contain(source, oth) }) {
		return
	}
	a.Fn(fmterror.Message{
		Method: "NotContain",
		Cause:  "Source contains the received value",
		Values: []fmterror.Value{
			{
				Label: "source",
				Value: source,
			},
			{
				Label: "other",
				Value: oth,
			},
		},
		UserMessage: msg,
	})
}

func (a Asserter) ContainExactly(expected, actual interface{}, msg ...interface{}) {
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

func (a Asserter) containExactlyMap(exp reflect.Value, act reflect.Value, msg []interface{}) {
	a.TB.Helper()

	if a.eq(exp.Interface(), act.Interface()) {
		return
	}
	a.Fn(fmterror.Message{
		Method: "ContainExactly",
		Cause:  "SubMap content doesn't exactly match with expectations.",
		Values: []fmterror.Value{
			{Label: "expected", Value: exp.Interface()},
			{Label: "actual", Value: act.Interface()},
		},
		UserMessage: msg,
	})
}

func (a Asserter) containExactlySlice(exp reflect.Value, act reflect.Value, msg []interface{}) {
	a.TB.Helper()

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
			a.Fn(fmterror.Message{
				Method: "ContainExactly",
				Cause:  fmt.Sprintf("Element not found at index %d", i),
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
				UserMessage: msg,
			})
		}
	}
}

func (a Asserter) AnyOf(blk func(a *AnyOf), msg ...interface{}) {
	a.TB.Helper()
	anyOf := &AnyOf{TB: a.TB, Fn: a.Fn}
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
func (a Asserter) Empty(v interface{}, msg ...interface{}) {
	a.TB.Helper()
	if a.isEmpty(v) {
		return
	}
	a.Fn(fmterror.Message{
		Method: "Empty",
		Cause:  "Value was expected to be empty.",
		Values: []fmterror.Value{
			{Label: "value", Value: v},
		},
		UserMessage: msg,
	})
}

// NotEmpty gets whether the specified value is considered empty.
func (a Asserter) NotEmpty(v interface{}, msg ...interface{}) {
	a.TB.Helper()
	if !a.try(func(a Asserter) { a.Empty(v) }) {
		return
	}
	a.Fn(fmterror.Message{
		Method: "NotEmpty",
		Cause:  "Value was expected to be not empty.",
		Values: []fmterror.Value{
			{Label: "value", Value: v},
		},
		UserMessage: msg,
	})
}

// ErrorIs allows you to assert an error value by an expectation.
// if the implementation of the test subject later changes, and for example, it starts to use wrapping,
// this should not be an issue as the IsEqualErr's error chain is also matched against the expectation.
func (a Asserter) ErrorIs(expected, actual error, msg ...interface{}) {
	a.TB.Helper()

	if errors.Is(actual, expected) {
		return
	}
	if a.eq(expected, actual) {
		return
	}
	a.Fn(fmterror.Message{
		Method: "ErrorIs",
		Cause:  "The actual error is not what was expected.",
		Values: []fmterror.Value{
			{Label: "expected", Value: expected},
			{Label: "actual", Value: actual},
		},
		UserMessage: msg,
	})
}
