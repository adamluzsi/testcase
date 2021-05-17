package spechelper

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/internal"
)

func OrderAsDefined(tb testing.TB) {
	internal.SetupCacheFlush(tb)
	testcase.SetEnv(tb, testcase.EnvKeyOrdering, string(testcase.OrderingAsDefined))
}
