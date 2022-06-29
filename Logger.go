package testcase

import "fmt"

func WithLogger(l Var[Logger]) SpecOption {
	return specOptionFunc(func(s *Spec) {
		l.Bind(s)
		s.logger = &l
	})
}

type Logger interface {
	Log(args ...any)
	Error(args ...any)
}

func (t *T) logger() Logger {
	t.TB.Helper()
	var l Logger = t.TB
	for _, spec := range t.spec.specsFromCurrent() {
		if spec.logger != nil {
			l = spec.logger.Get(t)
			break
		}
	}
	return l
}

func (t *T) log(log func(...any), args []any) {
	t.TB.Helper()
	t.logf(log, "%s", []any{fmt.Sprintln(args...)})
}

func (t *T) logf(log func(...any), format string, args []any) {
	t.TB.Helper()
	log(fmt.Sprintf(format, args...))
}

func (t *T) Log(args ...any) {
	t.TB.Helper()
	t.log(t.logger().Log, args)
}

func (t *T) Logf(format string, args ...any) {
	t.TB.Helper()
	t.logf(t.logger().Log, format, args)
}

func (t *T) Error(args ...any) {
	t.TB.Helper()
	t.log(t.logger().Error, args)
	t.Fail()
}

func (t *T) Errorf(format string, args ...any) {
	t.TB.Helper()
	t.logf(t.logger().Error, format, args)
	t.Fail()
}

func (t *T) Fatal(args ...any) {
	t.TB.Helper()
	t.log(t.logger().Error, args)
	t.FailNow()
}

func (t *T) Fatalf(format string, args ...any) {
	t.TB.Helper()
	t.logf(t.logger().Error, format, args)
	t.FailNow()
}

func (t *T) Skip(args ...any) {
	t.TB.Helper()
	t.log(t.logger().Log, args)
	t.SkipNow()
}

func (t *T) Skipf(format string, args ...any) {
	t.TB.Helper()
	t.logf(t.logger().Log, format, args)
	t.SkipNow()
}
