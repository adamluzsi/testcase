package internal_test

import (
	"runtime"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal"
)

func TestInGoroutine(t *testing.T) {
	var total int
	var hasRun bool
	var survived bool
	defer func() { assert.Must(t).True(survived) }()
	internal.InGoroutine(func() {
		total++
		hasRun = true
		runtime.Goexit()
	})
	survived = true
	assert.Must(t).Equal(1, total)
	assert.Must(t).True(hasRun)
}

func TestInGoroutine_panic(t *testing.T) {
	panicValue := func() (panicValue interface{}) {
		defer func() { panicValue = recover() }()
		internal.InGoroutine(func() { panic(`boom`) })
		return nil
	}()
	//
	assert.Must(t).Equal(`boom`, panicValue)
}
