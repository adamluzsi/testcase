package internal_test

import (
	"runtime"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/adamluzsi/testcase/internal"
	"github.com/stretchr/testify/require"
)

func TestStubTB(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		stub = s.Let(`stub`, func(t *testcase.T) interface{} {
			return &internal.StubTB{}
		})
		stubGet = func(t *testcase.T) *internal.StubTB {
			return stub.Get(t).(*internal.StubTB)
		}
	)

	s.Test(`.Cleanup + .Finish`, func(t *testcase.T) {
		var i int
		stubGet(t).Cleanup(func() { i++ })
		stubGet(t).Cleanup(func() { i++ })
		stubGet(t).Cleanup(func() { i++ })
		require.Equal(t, 0, i)
		stubGet(t).Finish()
		require.Equal(t, 3, i)
	})

	s.Test(`.Cleanup + .Finish + runtime.Goexit`, func(t *testcase.T) {
		var i int
		stubGet(t).Cleanup(func() { runtime.Goexit() })
		stubGet(t).Cleanup(func() { i++ })
		stubGet(t).Cleanup(func() { i++ })
		require.Equal(t, 0, i)
		stubGet(t).Finish()
		require.Equal(t, 2, i)
	})

	s.Test(`.Error`, func(t *testcase.T) {
		require.False(t, stubGet(t).IsFailed)
		stubGet(t).Error(`arg1`, `arg2`, `arg3`)
		require.True(t, stubGet(t).IsFailed)
	})

	s.Test(`.Errorf`, func(t *testcase.T) {
		require.False(t, stubGet(t).IsFailed)
		stubGet(t).Errorf(`%s %s %s`, `arg1`, `arg2`, `arg3`)
		require.True(t, stubGet(t).IsFailed)
	})

	s.Test(`.Fail`, func(t *testcase.T) {
		require.False(t, stubGet(t).IsFailed)
		stubGet(t).Fail()
		require.True(t, stubGet(t).IsFailed)
	})

	s.Test(`.FailNow`, func(t *testcase.T) {
		require.False(t, stubGet(t).IsFailed)
		var ran bool
		internal.InGoroutine(func() {
			stubGet(t).FailNow()
			ran = true
		})
		require.False(t, ran)
		require.True(t, stubGet(t).IsFailed)
	})

	s.Test(`.Failed`, func(t *testcase.T) {
		require.False(t, stubGet(t).Failed())
		stubGet(t).Fail()
		require.True(t, stubGet(t).Failed())
	})

	s.Test(`.Fatal`, func(t *testcase.T) {
		require.False(t, stubGet(t).IsFailed)
		var ran bool
		internal.InGoroutine(func() {
			stubGet(t).Fatal(`arg1`, `arg2`, `arg3`)
			ran = true
		})
		require.False(t, ran)
		require.True(t, stubGet(t).IsFailed)
	})

	s.Test(`.Fatalf`, func(t *testcase.T) {
		require.False(t, stubGet(t).IsFailed)
		var ran bool
		internal.InGoroutine(func() {
			stubGet(t).Fatalf(`%s %s %s`, `arg1`, `arg2`, `arg3`)
			ran = true
		})
		require.False(t, ran)
		require.True(t, stubGet(t).IsFailed)
	})

	s.Test(`.Helper`, func(t *testcase.T) {
		stubGet(t).Helper()
	})

	s.Test(`.Log`, func(t *testcase.T) {
		stubGet(t).Log()
	})

	s.Test(`.Logf`, func(t *testcase.T) {
		stubGet(t).Logf(`%s %s %s`, `arg1`, `arg2`, `arg3`)
	})

	s.Test(`.Name`, func(t *testcase.T) {
		val := fixtures.Random.String()
		stubGet(t).StubName = val
		require.Equal(t, val, stubGet(t).Name())
	})

	s.Test(`.Skip`, func(t *testcase.T) {
		require.False(t, stubGet(t).Skipped())
		var ran bool
		internal.InGoroutine(func() {
			stubGet(t).Skip()
			ran = true
		})
		require.False(t, ran)
		require.True(t, stubGet(t).Skipped())
	})

	s.Test(`.SkipNow + .Skipped`, func(t *testcase.T) {
		require.False(t, stubGet(t).Skipped())
		var ran bool
		internal.InGoroutine(func() {
			stubGet(t).SkipNow()
			ran = true
		})
		require.False(t, ran)
		require.True(t, stubGet(t).Skipped())
	})

	s.Test(`.Skipf`, func(t *testcase.T) {
		require.False(t, stubGet(t).Skipped())
		var ran bool
		internal.InGoroutine(func() {
			stubGet(t).Skipf(`%s`, `arg42`)
			ran = true
		})
		require.False(t, ran)
		require.True(t, stubGet(t).Skipped())
	})

	s.Test(`.TempDir`, func(t *testcase.T) {
		val := fixtures.Random.String()
		stubGet(t).StubTempDir = val
		require.Equal(t, val, stubGet(t).TempDir())
	})
}
