package testcase

import (
	"strings"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal"
)

func TestSpec_FriendlyVarNotDefined(t *testing.T) {
	stub := &internal.StubTB{}
	s := NewSpec(stub)
	willFatalWithMessage := willFatalWithMessageFn(stub)

	v1 := Let[string](s, func(t *T) string { return `hello-world` })
	v2 := Let[string](s, func(t *T) string { return `hello-world` })
	tct := NewT(stub, s)

	s.Test(`var1 var found`, func(t *T) {
		assert.Must(t).Equal(`hello-world`, v1.Get(t))
	})

	t.Run(`not existing var will panic with friendly msg`, func(t *testing.T) {
		panicMSG := willFatalWithMessage(t, func() { tct.I(`not-exist`) })
		msg := strings.Join(panicMSG, " ")
		assert.Must(t).Contain(msg, `Variable "not-exist" is not found`)
		assert.Must(t).Contain(msg, `Did you mean?`)
		assert.Must(t).Contain(msg, v1.ID)
		assert.Must(t).Contain(msg, v2.ID)
	})
}

func isFatalFn(stub *internal.StubTB) func(block func()) bool {
	return func(block func()) bool {
		stub.IsFailed = false
		defer func() { stub.IsFailed = false }()
		var finished bool
		internal.RecoverExceptGoexit(func() {
			block()
			finished = true
		})
		return !finished && stub.Failed()
	}
}

func willFatalWithMessageFn(stub *internal.StubTB) func(tb testing.TB, blk func()) []string {
	isFatal := isFatalFn(stub)
	return func(tb testing.TB, blk func()) []string {
		stub.Logs = nil
		assert.Must(tb).True(isFatal(blk))
		return stub.Logs
	}
}
