package testcase_test

import (
	"bytes"
	"testing"

	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/sandbox"

	"github.com/adamluzsi/testcase/assert"
)

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
	tb.Helper()
	tb.Skip(`unsupported`)
}

func isFatalFn(stub *doubles.TB) func(block func()) bool {
	return func(block func()) bool {
		stub.IsFailed = false
		defer func() { stub.IsFailed = false }()

		var finished bool
		sandbox.Run(func() {
			block()
			finished = true
		})

		ltb, ok := stub.LastRunTB()
		if !ok {
			ltb = stub
		}

		return !finished && (ltb.Failed() || stub.Failed())
	}
}

func willFatalWithMessageFn(stub *doubles.TB) func(tb testing.TB, blk func()) string {
	isFatal := isFatalFn(stub)
	return func(tb testing.TB, blk func()) string {
		stub.Logs = bytes.Buffer{}
		assert.Must(tb).True(isFatal(blk))
		return stub.Logs.String()
	}
}
