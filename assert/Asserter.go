package assert

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type Asserter struct {
	Helper func()
	FailFn func(args ...interface{})
}

func (a Asserter) True(v bool, msg ...interface{}) {
	a.Helper()

	if !v {
		a.FailFn(message{
			Method: "True",
			Cause:  "",
			Left: &messageValue{
				Label: "value",
				Value: v,
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
	a.FailFn(message{
		Method: "Nil",
		Cause:  "Not nil value received",
		Left: &messageValue{
			Label: "value",
			Value: v,
		},
		UserMessage: msg,
	})
}

func (a Asserter) NotNil(v interface{}, msg ...interface{}) {
	a.Helper()
	if v != nil {
		return
	}
	a.FailFn(message{
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
	a.FailFn(message{
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
	a.FailFn(message{
		Method: "Panics",
		Cause:  "Expected to panic or die.",
		Left: &messageValue{
			Label: "panic:",
			Value: panicValue,
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

func (a Asserter) eq(exp, act interface{}) bool {
	return reflect.DeepEqual(exp, act)
}

func (a Asserter) failEqual(expected interface{}, actually interface{}, msg []interface{}) {
	a.Helper()

	a.FailFn(message{
		Method: "Equal",
		Left: &messageValue{
			Label: "expected",
			Value: expected,
		},
		Right: &messageValue{
			Label: "actual",
			Value: actually,
		},
		UserMessage: msg,
	}.String())
}

func (a Asserter) mustBeEquable(vs ...interface{}) bool {
	a.Helper()

	fail := func(v interface{}) bool {
		a.FailFn(message{
			Method: "Equal",
			Cause:  "Value is expected to be equable.",
			Left: &messageValue{
				Label: "value",
				Value: v,
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
		a.FailFn(message{
			Method: "Contains",
			Cause:  "invalid source value",
			Left: &messageValue{
				Label: "value",
				Value: src,
			},
		}.String())
		return
	}
	if !rHas.IsValid() {
		a.FailFn(message{
			Method: "Contains",
			Cause:  `invalid "has" value`,
			Left: &messageValue{
				Label: "value",
				Value: has,
			},
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
		panic(message{
			Method: "Contains",
			Cause:  "unimplemented scenario",
			Left: &messageValue{
				Label: "source",
				Value: fmt.Sprintf("%T", src),
			},
			Right: &messageValue{
				Label: "has",
				Value: fmt.Sprintf("%T", has),
			},
		}.String())
	}
}

func (a Asserter) failContains(src, sub interface{}, msg ...interface{}) {
	a.Helper()

	a.FailFn(message{
		Method: "Contains",
		Cause:  "Source doesn't contains expected value(s).",
		Left: &messageValue{
			Label: "source",
			Value: src,
		},
		Right: &messageValue{
			Label: "sub",
			Value: sub,
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
	a.FailFn(message{
		Method: "Contains",
		Cause:  "Couldn't find the expected value in the source slice",
		Left: &messageValue{
			Label: "source",
			Value: slice.Interface(),
		},
		Right: &messageValue{
			Label: "value",
			Value: value.Interface(),
		},
		UserMessage: msg,
	})
}

func (a Asserter) sliceContainsSubSlice(slice, sub reflect.Value, msg []interface{}) {
	a.Helper()

	failWithNotEqual := func() { a.failContains(slice.Interface(), sub.Interface(), msg...) }

	if slice.Kind() != reflect.Slice || sub.Kind() != reflect.Slice {
		a.FailFn(message{
			Method: "Contains",
			Cause:  "Invalid slice type(s).",
			Left: &messageValue{
				Label: "source",
				Value: slice.Interface(),
			},
			Right: &messageValue{
				Label: "sub",
				Value: sub.Interface(),
			},
			UserMessage: msg,
		}.String())
		return
	}
	if slice.Len() < sub.Len() {
		a.FailFn(message{
			Method: "Contains",
			Cause:  "Source slice is smaller than sub slice.",
			Left: &messageValue{
				Label: "source",
				Value: slice.Interface(),
			},
			Right: &messageValue{
				Label: "sub",
				Value: sub.Interface(),
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
			a.FailFn(message{
				Method: "Contains",
				Cause:  "Source doesn't contains the other map.",
				Left: &messageValue{
					Label: "source",
					Value: src.Interface(),
				},
				Right: &messageValue{
					Label: "key",
					Value: key.Interface(),
				},
				UserMessage: msg,
			})
			return
		}
		if !a.eq(srcValue.Interface(), has.MapIndex(key).Interface()) {
			a.FailFn(message{
				Method: "Contains",
				Cause:  "Source has the key but with different value.",
				Left: &messageValue{
					Label: "source",
					Value: src.Interface(),
				},
				Right: &messageValue{
					Label: "key",
					Value: key.Interface(),
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
	a.FailFn(message{
		Method: "Contains",
		Cause:  "String doesn't include sub string.",
		Left: &messageValue{
			Label: "string",
			Value: src.Interface(),
		},
		Right: &messageValue{
			Label: "substr",
			Value: has.Interface(),
		},
		UserMessage: msg,
	})
}

func (a Asserter) NotContain(source, oth interface{}, msg ...interface{}) {
	a.Helper()
	var failed bool
	Asserter{Helper: a.Helper, FailFn: func(args ...interface{}) { failed = true }}.Contain(source, oth)
	if failed {
		return
	}
	a.FailFn(message{
		Method: "NotContain",
		Cause:  "Source contains the received value",
		Left: &messageValue{
			Label: "source",
			Value: source,
		},
		Right: &messageValue{
			Label: "other",
			Value: oth,
		},
		UserMessage: msg,
	})
}

func (a Asserter) ContainExactly(expected, actual interface{}, msg ...interface{}) {
	a.Helper()

	exp := reflect.ValueOf(expected)
	act := reflect.ValueOf(actual)

	if !exp.IsValid() {
		panic(message{
			Method: "ContainExactly",
			Cause:  "invalid expected value",
			Left: &messageValue{
				Label: "value",
				Value: expected,
			},
		}.String())
	}
	if !act.IsValid() {
		panic(message{
			Method: "ContainExactly",
			Cause:  `invalid actual value`,
			Left: &messageValue{
				Label: "value",
				Value: actual,
			},
		}.String())
	}

	switch {
	case exp.Kind() == reflect.Slice && exp.Type() == act.Type():
		a.containExactlySlice(exp, act, msg)

	case exp.Kind() == reflect.Map && exp.Type() == act.Type():
		a.containExactlyMap(exp, act, msg)

	default:
		panic(message{
			Method: "ContainExactly",
			Cause:  "Unimplemented scenario / type mismatch.",
			Left: &messageValue{
				Label: "type of expected",
				Value: fmt.Sprintf("%T", expected),
			},
			Right: &messageValue{
				Label: "type of actual",
				Value: fmt.Sprintf("%T", actual),
			},
		}.String())
	}
}

func (a Asserter) containExactlyMap(exp reflect.Value, act reflect.Value, msg []interface{}) {
	a.Helper()

	if a.eq(exp.Interface(), act.Interface()) {
		return
	}
	a.FailFn(message{
		Method: "ContainExactly",
		Cause:  "SubMap content doesn't exactly match with expectations.",
		Left: &messageValue{
			Label: "expected",
			Value: exp.Interface(),
		},
		Right: &messageValue{
			Label: "actual",
			Value: act.Interface(),
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
			a.FailFn(message{
				Method: "ContainExactly",
				Cause:  fmt.Sprintf("Element not found at index %d", i),
				Left: &messageValue{
					Label: "actual:",
					Value: act.Interface(),
				},
				Right: &messageValue{
					Label: "value",
					Value: expectedValue,
				},
				UserMessage: msg,
			})
		}
	}
}
