package let

import (
	"context"
	"time"

	"github.com/adamluzsi/testcase"
)

func Context(s *testcase.Spec) testcase.Var[context.Context] {
	return testcase.Let(s, func(t *testcase.T) context.Context {
		return context.Background()
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
