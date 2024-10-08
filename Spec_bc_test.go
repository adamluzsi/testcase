package testcase

import (
	"bytes"
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/sandbox"
)

func TestSpec_FriendlyVarNotDefined(t *testing.T) {
	stub := &doubles.TB{}
	s := NewSpec(stub)
	willFatalWithMessage := willFatalWithMessageFn(stub)

	v1 := Let[string](s, func(t *T) string { return `hello-world` })
	v2 := Let[string](s, func(t *T) string { return `hello-world` })
	tct := NewTWithSpec(stub, s)

	s.Test(`var1 var found`, func(t *T) {
		assert.Must(t).Equal(`hello-world`, v1.Get(t))
	})

	t.Run(`not existing var will panic with friendly msg`, func(t *testing.T) {
		msg := willFatalWithMessage(t, func() { tct.vars.Get(tct, `not-exist`) })
		assert.Must(t).Contain(msg.String(), `Variable "not-exist" is not found`)
		assert.Must(t).Contain(msg.String(), `Did you mean?`)
		assert.Must(t).Contain(msg.String(), v1.ID)
		assert.Must(t).Contain(msg.String(), v2.ID)
	})
}

func isFatalFn(stub *doubles.TB) func(block func()) bool {
	return func(block func()) bool {
		stub.IsFailed = false
		defer func() { stub.IsFailed = false }()
		out := sandbox.Run(block)
		return !out.OK && stub.Failed()
	}
}

func willFatalWithMessageFn(stub *doubles.TB) func(tb testing.TB, blk func()) bytes.Buffer {
	isFatal := isFatalFn(stub)
	return func(tb testing.TB, blk func()) bytes.Buffer {
		stub.Logs = bytes.Buffer{}
		assert.Must(tb).True(isFatal(blk))
		return stub.Logs
	}
}
