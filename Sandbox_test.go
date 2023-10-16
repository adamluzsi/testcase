package testcase_test

import (
	"fmt"
	"runtime"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/random"
	"go.llib.dev/testcase/sandbox"
)

func ExampleSandbox() {
	stb := &doubles.TB{}
	outcome := testcase.Sandbox(func() {
		// some test helper function calls fatal, which cause runtime.Goexit after marking the test failed.
		stb.FailNow()
	})

	fmt.Println("The sandbox run has finished without an issue", outcome.OK)
	fmt.Println("runtime.Goexit was called:", outcome.Goexit)
	fmt.Println("panic value:", outcome.PanicValue)
}

func TestSandbox_smoke(t *testing.T) {
	var out sandbox.RunOutcome
	out = testcase.Sandbox(func() {})
	assert.True(t, out.OK)
	assert.Nil(t, out.PanicValue)

	out = testcase.Sandbox(func() { runtime.Goexit() })
	assert.False(t, out.OK)
	assert.Nil(t, out.PanicValue)

	expectedPanicValue := random.New(random.CryptoSeed{}).Error()
	out = testcase.Sandbox(func() { panic(expectedPanicValue) })
	assert.False(t, out.OK)
	assert.Equal[any](t, expectedPanicValue, out.PanicValue)
}
