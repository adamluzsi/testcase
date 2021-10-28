package internal_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/adamluzsi/testcase/internal"
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
		t.Must.Equal(0, i)
		stubGet(t).Finish()
		t.Must.Equal(3, i)
	})

	s.Test(`.Cleanup + .Finish + runtime.Goexit`, func(t *testcase.T) {
		var i int
		stubGet(t).Cleanup(func() { runtime.Goexit() })
		stubGet(t).Cleanup(func() { i++ })
		stubGet(t).Cleanup(func() { i++ })
		t.Must.Equal(0, i)
		stubGet(t).Finish()
		t.Must.Equal(2, i)
	})

	s.Test(`.Error`, func(t *testcase.T) {
		stb := stubGet(t)
		assert.Must(t).True(!stb.IsFailed)
		stb.Error(`arg1`, `arg2`, `arg3`)
		assert.Must(t).True(stb.IsFailed)
		t.Must.Contain(stb.Logs, fmt.Sprint(`arg1`, `arg2`, `arg3`))
	})

	s.Test(`.Errorf`, func(t *testcase.T) {
		stb := stubGet(t)
		assert.Must(t).True(!stb.IsFailed)
		stb.Errorf(`%s %s %s`, `arg1`, `arg2`, `arg3`)
		assert.Must(t).True(stb.IsFailed)
		t.Must.Contain(stb.Logs, fmt.Sprintf(`%s %s %s`, `arg1`, `arg2`, `arg3`))
	})

	s.Test(`.Fail`, func(t *testcase.T) {
		assert.Must(t).True(!stubGet(t).IsFailed)
		stubGet(t).Fail()
		assert.Must(t).True(stubGet(t).IsFailed)
	})

	s.Context(`.FailNow`, func(s *testcase.Spec) {
		s.Test(`by default it will exit the goroutine`, func(t *testcase.T) {
			assert.Must(t).True(!stubGet(t).IsFailed)
			var ran bool
			internal.RecoverExceptGoexit(func() {
				stubGet(t).FailNow()
				ran = true
			})
			assert.Must(t).True(!ran)
			assert.Must(t).True(stubGet(t).IsFailed)
		})

		s.Test(`when stubbed it will use the stubbed function`, func(t *testcase.T) {
			var stubFailNowRan bool
			stubGet(t).StubFailNow = func() { stubFailNowRan = true }
			assert.Must(t).True(!stubFailNowRan)
			assert.Must(t).True(!stubGet(t).IsFailed)
			var ran bool
			internal.RecoverExceptGoexit(func() {
				stubGet(t).FailNow()
				ran = true
			})
			assert.Must(t).True(ran)
			assert.Must(t).True(stubGet(t).IsFailed)
			assert.Must(t).True(stubFailNowRan)
		})
	})

	s.Test(`.FailNow`, func(t *testcase.T) {

	})

	s.Test(`.Failed`, func(t *testcase.T) {
		assert.Must(t).True(!stubGet(t).Failed())
		stubGet(t).Fail()
		assert.Must(t).True(stubGet(t).Failed())
	})

	s.Test(`.Fatal`, func(t *testcase.T) {
		stb := stubGet(t)
		assert.Must(t).True(!stb.IsFailed)
		var ran bool
		internal.RecoverExceptGoexit(func() {
			stb.Fatal(`arg1`, `arg2`, `arg3`)
			ran = true
		})
		assert.Must(t).True(!ran)
		assert.Must(t).True(stb.IsFailed)
		t.Must.Contain(stb.Logs, fmt.Sprint(`arg1`, `arg2`, `arg3`))
	})

	s.Test(`.Fatalf`, func(t *testcase.T) {
		assert.Must(t).True(!stubGet(t).IsFailed)
		var ran bool
		internal.RecoverExceptGoexit(func() {
			stubGet(t).Fatalf(`%s %s %s`, `arg1`, `arg2`, `arg3`)
			ran = true
		})
		assert.Must(t).True(!ran)
		assert.Must(t).True(stubGet(t).IsFailed)
		t.Must.Contain(stubGet(t).Logs, fmt.Sprintf(`%s %s %s`, `arg1`, `arg2`, `arg3`))
	})

	s.Test(`.Helper`, func(t *testcase.T) {
		stubGet(t).Helper()
	})

	s.Test(`.Log`, func(t *testcase.T) {
		stb := stubGet(t)
		stb.Log()
		t.Must.Equal(1, len(stb.Logs))
		t.Must.Contain(stb.Logs, "")
		stb.Log("foo", "bar", "baz")
		t.Must.Equal(2, len(stb.Logs))
		t.Must.Contain(stb.Logs, fmt.Sprint("foo", "bar", "baz"))
	})

	s.Test(`.Logf`, func(t *testcase.T) {
		stb := stubGet(t)
		stb.Logf(`%s %s %s`, `arg1`, `arg2`, `arg3`)
		t.Must.Equal(1, len(stb.Logs))
		t.Must.Contain(stb.Logs, fmt.Sprintf(`%s %s %s`, `arg1`, `arg2`, `arg3`))
		stb.Logf(`%s %s %s`, `arg4`, `arg5`, `arg6`)
		t.Must.Equal(2, len(stb.Logs))
		t.Must.Contain(stb.Logs, fmt.Sprintf(`%s %s %s`, `arg4`, `arg5`, `arg6`))
	})

	s.Test(`.Name`, func(t *testcase.T) {
		val := fixtures.Random.String()
		stubGet(t).StubName = val
		t.Must.Equal(val, stubGet(t).Name())
	})

	s.Test(`.Skip`, func(t *testcase.T) {
		assert.Must(t).True(!stubGet(t).Skipped())
		var ran bool
		internal.RecoverExceptGoexit(func() {
			stubGet(t).Skip()
			ran = true
		})
		assert.Must(t).True(!ran)
		assert.Must(t).True(stubGet(t).Skipped())
	})

	s.Test(`.SkipNow + .Skipped`, func(t *testcase.T) {
		assert.Must(t).True(!stubGet(t).Skipped())
		var ran bool
		internal.RecoverExceptGoexit(func() {
			stubGet(t).SkipNow()
			ran = true
		})
		assert.Must(t).True(!ran)
		assert.Must(t).True(stubGet(t).Skipped())
	})

	s.Test(`.Skipf`, func(t *testcase.T) {
		assert.Must(t).True(!stubGet(t).Skipped())
		var ran bool
		internal.RecoverExceptGoexit(func() {
			stubGet(t).Skipf(`%s`, `arg42`)
			ran = true
		})
		assert.Must(t).True(!ran)
		assert.Must(t).True(stubGet(t).Skipped())
	})

	s.Test(`.TempDir`, func(t *testcase.T) {
		val := fixtures.Random.String()
		stubGet(t).StubTempDir = val
		t.Must.Equal(val, stubGet(t).TempDir())
	})
}
