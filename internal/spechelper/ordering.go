package spechelper

import (
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/internal"
	"go.llib.dev/testcase/internal/environ"
)

func OrderAsDefined(tb testing.TB) {
	internal.SetupCacheFlush(tb)
	testcase.SetEnv(tb, environ.KeyOrdering, string(testcase.OrderingAsDefined))
}
