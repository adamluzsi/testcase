package assert_test

import (
	"strings"
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/random"
)

func ExampleMessage() {
	var tb testing.TB

	assert.True(tb, true, "this is a const which is interpreted as assertion.Message")
}

func TestMessage(t *testing.T) {
	dtb := &doubles.TB{}
	a := asserter(dtb)
	rnd := random.New(random.CryptoSeed{})
	exp := assert.Message(rnd.String())
	a.True(false, exp)
	assert.Contains(t, dtb.Logs.String(), strings.TrimSpace(string(exp)))
}

func TestMessagef(t *testing.T) {
	exp := assert.MessageF("answer:%d", 42)
	assert.Equal[assert.Message](t, exp, "answer:42")
}
