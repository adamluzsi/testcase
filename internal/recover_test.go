package internal_test

import (
	"runtime"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal"
)

func TestRecover(t *testing.T) {
	t.Run(`on no panic, provide information about it`, func(t *testing.T) {
		var hasRun bool
		panicValue, ok := internal.Recover(func() { hasRun = true })
		assert.Must(t).True(hasRun)
		assert.Must(t).True(ok)
		assert.Must(t).Nil(panicValue)
	})
	t.Run(`recovers from a panic`, func(t *testing.T) {
		const expected = "boom"
		var hasRun bool
		var survived bool
		defer func() { assert.Must(t).True(survived) }()
		internal.Recover(func() {
			hasRun = true
			panic(expected)
		})
		survived = true
		assert.Must(t).True(hasRun)
	})
	t.Run(`recovers from a goexit`, func(t *testing.T) {
		const expected = "boom"
		var hasRun bool
		var survived bool
		defer func() { assert.Must(t).True(survived) }()
		internal.Recover(func() {
			hasRun = true
			runtime.Goexit()
		})
		survived = true
		assert.Must(t).True(hasRun)
	})
	t.Run("let us know if there was a panic value", func(t *testing.T) {
		const expected = "boom"
		panicValue, ok := internal.Recover(func() { panic(expected) })
		assert.Must(t).True(!ok)
		assert.Must(t).Equal(expected, panicValue)
	})
	t.Run("no value on goexit", func(t *testing.T) {
		panicValue, ok := internal.Recover(func() { runtime.Goexit() })
		assert.Must(t).True(!ok)
		assert.Must(t).Nil(panicValue)
	})
}

func TestRecoverExceptGoexit(t *testing.T) {
	var total int
	var hasRun bool
	var survived bool
	defer func() { assert.Must(t).True(survived) }()
	internal.RecoverExceptGoexit(func() {
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
		internal.RecoverExceptGoexit(func() { panic(`boom`) })
		return nil
	}()
	//
	assert.Must(t).Equal(`boom`, panicValue)
}
