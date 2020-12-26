package testcase_test

import (
	"github.com/adamluzsi/testcase"
	"testing"
)

func ExampleSetEnv() {
	var tb testing.TB
	testcase.SetEnv(tb, `MY_KEY`, `myvalue`)
}
