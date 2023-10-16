package spechelper

import (
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/internal"
)

func OrderAsDefined(tb testing.TB) {
	internal.SetupCacheFlush(tb)
	testcase.SetEnv(tb, testcase.EnvKeyOrdering, string(testcase.OrderingAsDefined))
}
