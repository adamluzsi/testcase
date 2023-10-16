package internal_test

import (
	"runtime"
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/sandbox"
)

func TestRecoverFromGoexit(t *testing.T) {
	var total int
	var hasRun bool
	var survived bool
	defer func() { assert.Must(t).True(survived) }()
	sandbox.Run(func() {
		total++
		hasRun = true
		runtime.Goexit()
	})
	survived = true
	assert.Must(t).Equal(1, total)
	assert.Must(t).True(hasRun)
}
