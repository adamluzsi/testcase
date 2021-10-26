package assert

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/adamluzsi/testcase/internal/fmterror"
)

type Asserter struct {
	Helper func()
	FailFn func(args ...interface{})
}

func (a Asserter) try(blk func(a Asserter)) (ok bool) {
	var failed bool
	blk(Asserter{Helper: a.Helper, FailFn: func(args ...interface{}) { failed = true }})
	return !failed
}

func (a Asserter) True(v bool, msg ...interface{}) {
	a.Helper()

	if !v {
		a.FailFn(fmterror.Message{
			Method: "True",
			Cause:  "",
			Values: []fmterror.Value{
				{
					Label: "value",
					Value: v,
				},
			},
			UserMessage: msg,
		}.String())
		return
	}
}

func (a Asserter) Nil(v interface{}, msg ...interface{}) {
	a.Helper()
	if v == nil {
		return
	}
	if func() (isNil bool) {
		defer func() { _ = recover() }()

		return reflect.ValueOf(v).IsNil()
	}() {
		return
	}
	a.FailFn(fmterror.Message{
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
	a.Helper()
	if !a.try(func(a Asserter) { a.Nil(v) }) {
		return
	}
	a.FailFn(fmterror.Message{
		Method:      "NotNil",
		Cause:       "Nil value received",
		UserMessage: msg,
	})
}

func (a Asserter) hasPanicked(blk func()) (panicValue interface{}, ok bool) {
	a.Helper()
	var wg sync.WaitGroup
	wg.Add(1)
	var finished bool
	go func() {
		a.Helper()
		defer wg.Done()
		defer func() { panicValue = recover() }()
		blk()
		finished = true
	}()
	wg.Wait()
	return panicValue, !finished
}

func (a Asserter) Panic(blk func(), msg ...interface{}) (panicValue interface{}) {
	a.Helper()
	panicValue, ok := a.hasPanicked(blk)
	if ok {
		return panicValue
	}
	a.FailFn(fmterror.Message{
		Method:      "Panics",
		Cause:       "Expected to panic or die.",
		UserMessage: msg,
	})
	return nil
}

func (a Asserter) NotPanic(blk func(), msg ...interface{}) {
	a.Helper()
	panicValue, ok := a.hasPanicked(blk)
	if !ok {
		return
	}
	a.FailFn(fmterror.Message{
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

func (a Asserter) Equal(expected, actually interface{}, msg ...interface{}) {
	a.Helper()

	if !a.mustBeEquable(expected, actually) {
		return
	}

	// bytes.Equal(expected, actually)

	if !reflect.DeepEqual(expected, actually) {
		a.failEqual(expected, actually, msg)
		return
	}
}

func (a Asserter) NotEqual(v, oth interface{}, msg ...interface{}) {
	a.Helper()
	if !a.try(func(a Asserter) { a.Equal(v, oth) }) {
		return
	}
	a.FailFn(fmterror.Message{
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
	})
}

func (a Asserter) eq(exp, act interface{}) bool {
	return reflect.DeepEqual(exp, act)
}

func (a Asserter) failEqual(expected interface{}, actually interface{}, msg []interface{}) {
	a.Helper()

	a.FailFn(fmterror.Message{
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

func (a Asserter) mustBeEquable(vs ...interface{}) bool {
	a.Helper()

	fail := func(v interface{}) bool {
		a.FailFn(fmterror.Message{
			Method: "Equal",
			Cause:  "Value is expected to be equable.",
			Values: []fmterror.Value{
				{
					Label: "value",
					Value: v,
				},
			},
		}.String())
		return false
	}
	for _, v := range vs {
		if v == nil {
			continue
		}

		if reflect.TypeOf(v).Kind() == reflect.Func {
			return fail(v)
		}
	}
	return true
}

func (a Asserter) Contain(src, has interface{}, msg ...interface{}) {
	a.Helper()
	rSrc := reflect.ValueOf(src)
	rHas := reflect.ValueOf(has)
	if !rSrc.IsValid() {
		a.FailFn(fmterror.Message{
			Method: "Contains",
			Cause:  "invalid source value",
			Values: []fmterror.Value{
				{Label: "value", Value: src},
			},
		}.String())
		return
	}
	if !rHas.IsValid() {
		a.FailFn(fmterror.Message{
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
	a.Helper()

	a.FailFn(fmterror.Message{
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
	a.Helper()
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
	a.FailFn(fmterror.Message{
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
	a.Helper()

	failWithNotEqual := func() { a.failContains(slice.Interface(), sub.Interface(), msg...) }

	if slice.Kind() != reflect.Slice || sub.Kind() != reflect.Slice {
		a.FailFn(fmterror.Message{
			Method: "Contains",
			Cause:  "Invalid slice type(s).",
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
	if slice.Len() < sub.Len() {
		a.FailFn(fmterror.Message{
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
			if reflect.DeepEqual(slice.Index(i).Interface(), sub.Index(j).Interface()) {
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

		if !reflect.DeepEqual(expected, actual) {
			failWithNotEqual()
			return
		}
	}
}

func (a Asserter) mapContainsSubMap(src reflect.Value, has reflect.Value, msg []interface{}) {
	for _, key := range has.MapKeys() {
		srcValue := src.MapIndex(key)
		if !srcValue.IsValid() {
			a.FailFn(fmterror.Message{
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
			a.FailFn(fmterror.Message{
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
	a.Helper()
	if strings.Contains(fmt.Sprint(src.Interface()), fmt.Sprint(has.Interface())) {
		return
	}
	a.FailFn(fmterror.Message{
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
	a.Helper()
	if !a.try(func(a Asserter) { a.Contain(source, oth) }) {
		return
	}
	a.FailFn(fmterror.Message{
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
	a.Helper()

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
	a.Helper()

	if a.eq(exp.Interface(), act.Interface()) {
		return
	}
	a.FailFn(fmterror.Message{
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
	a.Helper()

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
			a.FailFn(fmterror.Message{
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
