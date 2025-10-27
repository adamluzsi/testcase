package let

import (
	"context"
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/internal"
	"go.llib.dev/testcase/pkg/synctest"
	"go.llib.dev/testcase/random"
)

// Var creates a stateless specification variable, serving as a blueprint for test cases to construct test runtime values.
// It also functions as both a getter and setter for the value associated to the test runtime.
func Var[V any](s *testcase.Spec, blk func(t *testcase.T) V) testcase.Var[V] {
	s.H().Helper()
	return testcase.Let(s, blk)
}

// Var2 creates a stateless specification variable, serving as a blueprint for test cases to construct test runtime values.
// It also functions as both a getter and setter for the value associated to the test runtime.
func Var2[V1, V2 any](s *testcase.Spec, blk func(t *testcase.T) (V1, V2)) (testcase.Var[V1], testcase.Var[V2]) {
	s.H().Helper()
	return testcase.Let2(s, blk)
}

// Var3 creates a stateless specification variable, serving as a blueprint for test cases to construct test runtime values.
// It also functions as both a getter and setter for the value associated to the test runtime.
func Var3[V1, V2, V3 any](s *testcase.Spec, blk func(t *testcase.T) (V1, V2, V3)) (testcase.Var[V1], testcase.Var[V2], testcase.Var[V3]) {
	s.H().Helper()
	return testcase.Let3(s, blk)
}

// VarOf is a shorthand for defining a testcase.Var[V] using an immutable value.
// So the function blocks can be skipped, which makes tests more readable.
func VarOf[V any](s *testcase.Spec, v V) testcase.Var[V] {
	s.H().Helper()
	return testcase.LetValue(s, v)
}

// Act is a syntax shortcut that improves auto-completion in code editors like VS Code or IntelliJ IDEA.
// It represents an immutable testing act, where the closure retrieves input argument variables.
// This ensures that the test scenario properly arranges the variables beforehand since Act itself remains immutable.
func Act[A any](fn func(t *testcase.T) A) func(t *testcase.T) A {
	return fn
}

// Act0 is a syntax shortcut that improves auto-completion in code editors like VS Code or IntelliJ IDEA.
// It represents an immutable testing action, where the closure retrieves input argument variables.
// This ensures that the test scenario properly arranges the variables beforehand since Act itself remains immutable.
func Act0(fn func(t *testcase.T)) func(t *testcase.T) {
	return fn
}

// Act1 is a syntax shortcut that improves auto-completion in code editors like VS Code or IntelliJ IDEA.
// It represents an immutable testing act, where the closure retrieves input argument variables.
// This ensures that the test scenario properly arranges the variables beforehand since Act itself remains immutable.
func Act1[A any](fn func(t *testcase.T) A) func(t *testcase.T) A {
	return fn
}

// Act2 is a syntax shortcut that improves auto-completion in code editors like VS Code or IntelliJ IDEA.
// It represents an immutable testing act, where the closure retrieves input argument variables.
// This ensures that the test scenario properly arranges the variables beforehand since Act itself remains immutable.
func Act2[A, B any](fn func(t *testcase.T) (A, B)) func(t *testcase.T) (A, B) {
	return fn
}

// Act3 is a syntax shortcut that improves auto-completion in code editors like VS Code or IntelliJ IDEA.
// It represents an immutable testing act, where the closure retrieves input argument variables.
// This ensures that the test scenario properly arranges the variables beforehand since Act itself remains immutable.
func Act3[A, B, C any](fn func(t *testcase.T) (A, B, C)) func(t *testcase.T) (A, B, C) {
	return fn
}

func With[V any, FN withFN[V]](s *testcase.Spec, fn FN) testcase.Var[V] {
	s.H().Helper()
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

func As[To, From any](Var testcase.Var[From]) testcase.VarGetter[To] {
	var (
		fromType = reflect.TypeOf((*From)(nil)).Elem()
		toType   = reflect.TypeOf((*To)(nil)).Elem()
	)
	if !fromType.ConvertibleTo(toType) {
		panic(fmt.Sprintf("you can't have %s as %s", fromType.String(), toType.String()))
	}
	return internal.VarGetterFunc[testcase.T, To](func(t *testcase.T) To {
		var rFrom = reflect.ValueOf(Var.Get(t))
		return rFrom.Convert(toType).Interface().(To)
	})
}

func Context(s *testcase.Spec) testcase.Var[context.Context] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) context.Context {
		ctx, cancel := context.WithCancel(context.Background())
		t.Defer(cancel)
		return ctx
	})
}

func ContextWithCancel(s *testcase.Spec) (testcase.Var[context.Context], testcase.Var[func()]) {
	s.H().Helper()
	return testcase.Let2(s, func(t *testcase.T) (context.Context, func()) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Defer(cancel)
		return ctx, cancel
	})
}

func Error(s *testcase.Spec) testcase.Var[error] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) error {
		return t.Random.Error()
	})
}

func String(s *testcase.Spec) testcase.Var[string] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.String()
	})
}

func StringNC(s *testcase.Spec, length int, charset string) testcase.Var[string] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.StringNC(length, charset)
	})
}

func HexN(s *testcase.Spec, length int) testcase.Var[string] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.HexN(length)
	})
}

func UUID(s *testcase.Spec) testcase.Var[string] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.UUID()
	})
}

func Bool(s *testcase.Spec) testcase.Var[bool] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) bool {
		return t.Random.Bool()
	})
}

func Int(s *testcase.Spec) testcase.Var[int] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) int {
		return t.Random.Int()
	})
}

func IntN(s *testcase.Spec, n int) testcase.Var[int] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) int {
		return t.Random.IntN(n)
	})
}

func IntB(s *testcase.Spec, min, max int) testcase.Var[int] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) int {
		return t.Random.IntBetween(min, max)
	})
}

func Time(s *testcase.Spec) testcase.Var[time.Time] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) time.Time {
		return t.Random.Time()
	})
}

func TimeB(s *testcase.Spec, from, to time.Time) testcase.Var[time.Time] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) time.Time {
		return t.Random.TimeBetween(from, to)
	})
}

func OneOf[V any](s *testcase.Spec, vs ...V) testcase.Var[V] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) V {
		return t.Random.Pick(vs).(V)
	})
}

func DurationBetween(s *testcase.Spec, min, max time.Duration) testcase.Var[time.Duration] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) time.Duration {
		return t.Random.DurationBetween(min, max)
	})
}

func Contact(s *testcase.Spec, opts ...internal.ContactOption) testcase.Var[random.Contact] {
	s.H().Helper()
	return testcase.Let[random.Contact](s, func(t *testcase.T) random.Contact {
		return t.Random.Contact(opts...)
	})
}

func FirstName(s *testcase.Spec, opts ...internal.ContactOption) testcase.Var[string] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.Contact(opts...).FirstName
	})
}

func LastName(s *testcase.Spec) testcase.Var[string] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.Contact().LastName
	})
}

func Email(s *testcase.Spec) testcase.Var[string] {
	s.H().Helper()
	return testcase.Let(s, func(t *testcase.T) string {
		return t.Random.Contact().Email
	})
}

func HTTPTestResponseRecorder(s *testcase.Spec) testcase.Var[*httptest.ResponseRecorder] {
	return testcase.Let(s, func(t *testcase.T) *httptest.ResponseRecorder {
		return httptest.NewRecorder()
	})
}

// Latch is a simple tool to help you coordinate goroutines in tests.
// It lets them wait for a signal before continuing,
// and automatically releases when the test ends;
// so no goroutines are left hanging, even if the test finishes earlier than it would utilise the Latch.
//
// To release the Latch:
//
//	latch.Get(t).Release
//	close(latch.Get(t))
//
// To have goroutines waiting on the latch:
//
//	<-latch.Get(t)
//	latch.Get(t).Wait()
//	<-latch.Get(t).Done()
func Phaser(s *testcase.Spec) testcase.Var[*synctest.Phaser] {
	return testcase.Let(s, func(t *testcase.T) *synctest.Phaser {
		var p synctest.Phaser
		t.Cleanup(p.Finish)
		return &p
	})
}
