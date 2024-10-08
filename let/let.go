package let

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/internal"
	"go.llib.dev/testcase/random"
)

func With[V any, FN withFN[V]](s *testcase.Spec, fn FN) testcase.Var[V] {
	var init testcase.VarInit[V]
	switch fnv := any(fn).(type) {
	case func() V:
		init = func(t *testcase.T) V { return fnv() }
	case func(testing.TB) V:
		init = func(t *testcase.T) V { return fnv(t) }
	case func(*testcase.T) V:
		init = fnv
	}
	return testcase.Let(s, init)
}

type withFN[V any] interface {
	func() V |
		func(testing.TB) V |
		func(*testcase.T) V
}

func As[To, From any](Var testcase.Var[From]) testcase.Var[To] {
	asID++
	fromType := reflect.TypeOf((*From)(nil)).Elem()
	toType := reflect.TypeOf((*To)(nil)).Elem()
	if !fromType.ConvertibleTo(toType) {
		panic(fmt.Sprintf("you can't have %s as %s", fromType.String(), toType.String()))
	}
	return testcase.Var[To]{
		ID: fmt.Sprintf("%s AS %T #%d", Var.ID, *new(To), asID),
		Init: func(t *testcase.T) To {
			var rFrom = reflect.ValueOf(Var.Get(t))
			return rFrom.Convert(toType).Interface().(To)
		},
	}
}

var asID int // adds extra safety that there won't be a name collision between two variables

func Context(s *testcase.Spec) testcase.Var[context.Context] {
	return testcase.Let(s, func(t *testcase.T) context.Context {
		ctx, cancel := context.WithCancel(context.Background())
		t.Defer(cancel)
		return ctx
	})
}

func ContextWithCancel(s *testcase.Spec) (testcase.Var[context.Context], testcase.Var[func()]) {
	return testcase.Let2(s, func(t *testcase.T) (context.Context, func()) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Defer(cancel)
		return ctx, cancel
	})
}

func Error(s *testcase.Spec) testcase.Var[error] {
	return testcase.Let(s, func(t *testcase.T) error {
		return t.Random.Error()
	})
}

func String(s *testcase.Spec) testcase.Var[string] {
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.String()
	})
}

func StringNC(s *testcase.Spec, length int, charset string) testcase.Var[string] {
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.StringNC(length, charset)
	})
}

func UUID(s *testcase.Spec) testcase.Var[string] {
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.UUID()
	})
}

func Bool(s *testcase.Spec) testcase.Var[bool] {
	return testcase.Let(s, func(t *testcase.T) bool {
		return t.Random.Bool()
	})
}

func Int(s *testcase.Spec) testcase.Var[int] {
	return testcase.Let(s, func(t *testcase.T) int {
		return t.Random.Int()
	})
}

func IntN(s *testcase.Spec, n int) testcase.Var[int] {
	return testcase.Let(s, func(t *testcase.T) int {
		return t.Random.IntN(n)
	})
}

func IntB(s *testcase.Spec, min, max int) testcase.Var[int] {
	return testcase.Let(s, func(t *testcase.T) int {
		return t.Random.IntBetween(min, max)
	})
}

func Time(s *testcase.Spec) testcase.Var[time.Time] {
	return testcase.Let(s, func(t *testcase.T) time.Time {
		return t.Random.Time()
	})
}

func TimeB(s *testcase.Spec, from, to time.Time) testcase.Var[time.Time] {
	return testcase.Let(s, func(t *testcase.T) time.Time {
		return t.Random.TimeBetween(from, to)
	})
}

func ElementFrom[V any](s *testcase.Spec, vs ...V) testcase.Var[V] {
	return testcase.Let(s, func(t *testcase.T) V {
		return t.Random.SliceElement(vs).(V)
	})
}

func DurationBetween(s *testcase.Spec, min, max time.Duration) testcase.Var[time.Duration] {
	return testcase.Let(s, func(t *testcase.T) time.Duration {
		return t.Random.DurationBetween(min, max)
	})
}

func Contact(s *testcase.Spec, opts ...internal.ContactOption) testcase.Var[random.Contact] {
	return testcase.Let[random.Contact](s, func(t *testcase.T) random.Contact {
		return t.Random.Contact(opts...)
	})
}

func FirstName(s *testcase.Spec, opts ...internal.ContactOption) testcase.Var[string] {
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.Contact(opts...).FirstName
	})
}

func LastName(s *testcase.Spec) testcase.Var[string] {
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.Contact().LastName
	})
}

func Email(s *testcase.Spec) testcase.Var[string] {
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.Contact().Email
	})
}
