package testcase_test

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal"
)

func TestStubTB(t *testing.T) {
	s := testcase.NewSpec(t)

	var stub = testcase.Let(s, func(t *testcase.T) *testcase.StubTB {
		return &testcase.StubTB{}
	})

	s.Test(`.Cleanup + .Finish`, func(t *testcase.T) {
		var i int
		stub.Get(t).Cleanup(func() { i++ })
		stub.Get(t).Cleanup(func() { i++ })
		stub.Get(t).Cleanup(func() { i++ })
		t.Must.Equal(0, i)
		stub.Get(t).Finish()
		t.Must.Equal(3, i)
	})

	s.Test(`.Cleanup + .Finish + runtime.Goexit`, func(t *testcase.T) {
		var i int
		stub.Get(t).Cleanup(func() { runtime.Goexit() })
		stub.Get(t).Cleanup(func() { i++ })
		stub.Get(t).Cleanup(func() { i++ })
		t.Must.Equal(0, i)
		stub.Get(t).Finish()
		t.Must.Equal(2, i)
	})

	s.Test(`.Error`, func(t *testcase.T) {
		stb := stub.Get(t)
		assert.Must(t).True(!stb.IsFailed)
		stb.Error(`arg1`, `arg2`, `arg3`)
		assert.Must(t).True(stb.IsFailed)
		t.Must.Contain(stb.Logs, fmt.Sprint(`arg1`, `arg2`, `arg3`))
	})

	s.Test(`.Errorf`, func(t *testcase.T) {
		stb := stub.Get(t)
		assert.Must(t).True(!stb.IsFailed)
		stb.Errorf(`%s %s %s`, `arg1`, `arg2`, `arg3`)
		assert.Must(t).True(stb.IsFailed)
		t.Must.Contain(stb.Logs, fmt.Sprintf(`%s %s %s`, `arg1`, `arg2`, `arg3`))
	})

	s.Test(`.Fail`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).IsFailed)
		stub.Get(t).Fail()
		assert.Must(t).True(stub.Get(t).IsFailed)
	})

	s.Test(`.FailNow`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).IsFailed)
		var ran bool
		internal.RecoverExceptGoexit(func() {
			stub.Get(t).FailNow()
			ran = true
		})
		assert.Must(t).True(!ran)
		assert.Must(t).True(stub.Get(t).IsFailed)
	})

	s.Test(`.Failed`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).Failed())
		stub.Get(t).Fail()
		assert.Must(t).True(stub.Get(t).Failed())
	})

	s.Test(`.Fatal`, func(t *testcase.T) {
		stb := stub.Get(t)
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
		assert.Must(t).True(!stub.Get(t).IsFailed)
		var ran bool
		internal.RecoverExceptGoexit(func() {
			stub.Get(t).Fatalf(`%s %s %s`, `arg1`, `arg2`, `arg3`)
			ran = true
		})
		assert.Must(t).True(!ran)
		assert.Must(t).True(stub.Get(t).IsFailed)
		t.Must.Contain(stub.Get(t).Logs, fmt.Sprintf(`%s %s %s`, `arg1`, `arg2`, `arg3`))
	})

	s.Test(`.Helper`, func(t *testcase.T) {
		stub.Get(t).Helper()
	})

	s.Test(`.Log`, func(t *testcase.T) {
		stb := stub.Get(t)
		stb.Log()
		t.Must.Equal(1, len(stb.Logs))
		t.Must.Contain(stb.Logs, "")
		stb.Log("foo", "bar", "baz")
		t.Must.Equal(2, len(stb.Logs))
		t.Must.Contain(stb.Logs, fmt.Sprint("foo", "bar", "baz"))
	})

	s.Test(`.Logf`, func(t *testcase.T) {
		stb := stub.Get(t)
		stb.Logf(`%s %s %s`, `arg1`, `arg2`, `arg3`)
		t.Must.Equal(1, len(stb.Logs))
		t.Must.Contain(stb.Logs, fmt.Sprintf(`%s %s %s`, `arg1`, `arg2`, `arg3`))
		stb.Logf(`%s %s %s`, `arg4`, `arg5`, `arg6`)
		t.Must.Equal(2, len(stb.Logs))
		t.Must.Contain(stb.Logs, fmt.Sprintf(`%s %s %s`, `arg4`, `arg5`, `arg6`))
	})

	s.Context(`.ID`, func(s *testcase.Spec) {
		s.Test(`with provided name, name is used`, func(t *testcase.T) {
			val := t.Random.String()
			stub.Get(t).StubName = val
			t.Must.Equal(val, stub.Get(t).Name())
		})

		s.Test(`without provided name, a name is created and consistently returned`, func(t *testcase.T) {
			stub.Get(t).StubName = ""
			t.Must.NotEmpty(stub.Get(t).Name())
			t.Must.Equal(stub.Get(t).Name(), stub.Get(t).Name())
		})
	})

	s.Test(`.Skip`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).Skipped())
		var ran bool
		internal.RecoverExceptGoexit(func() {
			stub.Get(t).Skip()
			ran = true
		})
		assert.Must(t).True(!ran)
		assert.Must(t).True(stub.Get(t).Skipped())
	})

	s.Test(`.Skip + args`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).Skipped())
		var ran bool
		args := []any{"Hello", "world!"}
		internal.RecoverExceptGoexit(func() {
			stub.Get(t).Skip(args...)
			ran = true
		})
		assert.Must(t).True(!ran)
		assert.Must(t).True(stub.Get(t).Skipped())
		assert.Must(t).Contain(strings.Join(stub.Get(t).Logs, "\n"), fmt.Sprint(args...))
	})

	s.Test(`.SkipNow + .Skipped`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).Skipped())
		var ran bool
		internal.RecoverExceptGoexit(func() {
			stub.Get(t).SkipNow()
			ran = true
		})
		assert.Must(t).True(!ran)
		assert.Must(t).True(stub.Get(t).Skipped())
	})

	s.Test(`.Skipf`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).Skipped())
		var ran bool
		internal.RecoverExceptGoexit(func() {
			stub.Get(t).Skipf(`%s`, `arg42`)
			ran = true
		})
		assert.Must(t).True(!ran)
		assert.Must(t).True(stub.Get(t).Skipped())
	})

	s.Context(`.TempDir`, func(s *testcase.Spec) {
		s.Test(`with provided temp dir value, value is returned`, func(t *testcase.T) {
			val := t.Random.String()
			stub.Get(t).StubTempDir = val
			t.Must.Equal(val, stub.Get(t).TempDir())
		})
		s.Test(`without a provided temp dir, os temp dir returned`, func(t *testcase.T) {
			t.Must.Equal(os.TempDir(), stub.Get(t).TempDir())
		})
		s.Test("with a testing.TB provided, testing.TB TempDir is created", func(t *testcase.T) {
			st := stub.Get(t)
			st.TB = t
			tmpDir := st.TempDir()

			stat, err := os.Stat(tmpDir)
			t.Must.Nil(err)
			t.Must.True(stat.IsDir())
		})
	})
}
