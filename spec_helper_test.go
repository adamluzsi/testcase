package testcase_test

import "testing"

type CustomTB struct {
	testing.TB
	isFatalFCalled bool
}

func (tb *CustomTB) Run(name string, blk func(tb testing.TB)) bool {
	switch tb := tb.TB.(type) {
	case *testing.T:
		return tb.Run(name, func(t *testing.T) { blk(t) })
	case *testing.B:
		return tb.Run(name, func(b *testing.B) { blk(b) })
	default:
		panic("implement me")
	}
}

func (t *CustomTB) Fatalf(format string, args ...interface{}) {
	t.isFatalFCalled = true
	return
}

func unsupported(tb testing.TB) {
	tb.Skip(`unsupported`)
}
