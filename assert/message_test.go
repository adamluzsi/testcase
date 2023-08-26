package assert_test

import (
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/random"
	"testing"
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
	assert.Contain(t, dtb.Logs.String(), string(exp))
}
