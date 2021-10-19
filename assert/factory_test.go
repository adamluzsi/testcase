package assert_test

import (
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal"
)

func TestMust(t *testing.T) {
	h := assert.Must(t)
	var failedNow bool
	stub := &internal.StubTB{StubFailNow: func() { failedNow = true }}
	a := assert.Must(stub)
	a.True(false) // fail it
	h.True(failedNow)
	h.True(stub.IsFailed)
}

func TestShould(t *testing.T) {
	h := assert.Must(t)
	var failedNow bool
	stub := &internal.StubTB{StubFailNow: func() { failedNow = true }}
	a := assert.Should(stub)
	a.True(false) // fail it
	h.True(!failedNow)
	h.True(stub.IsFailed)
}
