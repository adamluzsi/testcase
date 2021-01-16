package testcase_test

import (
	"github.com/adamluzsi/testcase"
	"testing"
)

func ExampleSetEnv() {
	var tb testing.TB
	testcase.SetEnv(tb, `MY_KEY`, `myvalue`)
	// env will be restored after the test
}

func ExampleUnsetEnv() {
	var tb testing.TB
	testcase.UnsetEnv(tb, `MY_KEY`)
	// env will be restored after the test
}
