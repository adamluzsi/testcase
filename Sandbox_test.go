package testcase_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/random"
	"github.com/adamluzsi/testcase/sandbox"
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
