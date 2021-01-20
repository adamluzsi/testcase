package spechelper

import (
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/internal"
	"testing"
)

func OrderAsDefined(tb testing.TB) {
	internal.DisableCache(tb)
	testcase.SetEnv(tb, testcase.EnvKeyOrderMod, string(testcase.OrderingAsDefined))
}
