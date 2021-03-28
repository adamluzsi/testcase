package internal_test

import (
	"runtime"
	"testing"

	"github.com/adamluzsi/testcase/internal"
	"github.com/stretchr/testify/require"
)

func TestInGoroutine(t *testing.T) {
	var total int
	var hasRun bool
	var survived bool
	defer func() { require.True(t, survived) }()
	internal.InGoroutine(func() {
		total++
		hasRun = true
		runtime.Goexit()
	})
	survived = true
	require.Equal(t, 1, total)
	require.True(t, hasRun)
}

func TestInGoroutine_panic(t *testing.T) {
	panicValue := func() (panicValue interface{}) {
		defer func() { panicValue = recover() }()
		internal.InGoroutine(func() { panic(`boom`) })
		return nil
	}()
	//
	require.Equal(t, `boom`, panicValue)
}
