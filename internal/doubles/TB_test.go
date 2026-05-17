package doubles_test

import (
	"context"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/random"
	"go.llib.dev/testcase/sandbox"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/contracts"
	"go.llib.dev/testcase/internal/doubles"
)

func TestTB(t *testing.T) {
	s := testcase.NewSpec(t)

	var stub = testcase.Let(s, func(t *testcase.T) *doubles.TB {
		return &doubles.TB{}
	})

	s.Test(`.Cleanup + .Finish`, func(t *testcase.T) {
		var i int
		stub.Get(t).Cleanup(func() { i++ })
		stub.Get(t).Cleanup(func() { i++ })
		stub.Get(t).Cleanup(func() { i++ })
		assert.Must(t).Equal(0, i)
		stub.Get(t).Finish()
		assert.Must(t).Equal(3, i)
	})

	s.Test(`.Cleanup + .Finish + runtime.Goexit`, func(t *testcase.T) {
		var i int
		stub.Get(t).Cleanup(func() { runtime.Goexit() })
		stub.Get(t).Cleanup(func() { i++ })
		stub.Get(t).Cleanup(func() { i++ })
		assert.Must(t).Equal(0, i)
		stub.Get(t).Finish()
		assert.Must(t).Equal(2, i)
	})

	s.Test(`.Error`, func(t *testcase.T) {
		stb := stub.Get(t)
		assert.Must(t).True(!stb.IsFailed)
		stb.Error(`arg1`, `arg2`, `arg3`)
		assert.Must(t).True(stb.IsFailed)
		assert.Must(t).Contains(stb.Logs.String(), "arg1 arg2 arg3\n")
	})

	s.Test(`.Errorf`, func(t *testcase.T) {
		stb := stub.Get(t)
		assert.Must(t).True(!stb.IsFailed)
		stb.Errorf(`%s %q %s`, `arg1`, `arg2`, `arg3`)
		assert.Must(t).True(stb.IsFailed)
		assert.Must(t).Contains(stb.Logs.String(), "arg1 \"arg2\" arg3\n")
	})

	s.Test(`.Fail`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).IsFailed)
		stub.Get(t).Fail()
		assert.Must(t).True(stub.Get(t).IsFailed)
	})

	s.Context(`.FailNow`, func(s *testcase.Spec) {

		s.Before(func(t *testcase.T) {
			assert.Must(t).True(!stub.Get(t).IsFailed)
		})

		s.Test("", func(t *testcase.T) {
			var ran bool
			sandbox.Run(func() {
				stub.Get(t).FailNow()
				ran = true
			})
			assert.Must(t).False(ran)
			assert.Must(t).True(stub.Get(t).IsFailed)
		})

		s.Test("", func(t *testcase.T) {
			var failNowRan bool
			stub.Get(t).OnFailNow = func() { failNowRan = true }

			var ran bool
			sandbox.Run(func() {
				stub.Get(t).FailNow()
				ran = true
			})
			assert.Must(t).False(ran)
			assert.Must(t).True(stub.Get(t).IsFailed)
			assert.Must(t).True(failNowRan)
		})
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
		sandbox.Run(func() {
			stb.Log("-")
			stb.Fatal(`arg1`, `arg2`, `arg3`)
			ran = true
		})
		assert.Must(t).False(ran)
		assert.Must(t).True(stb.IsFailed)
		assert.Must(t).Contains(stb.Logs.String(), "-\narg1 arg2 arg3\n")
	})

	s.Test(`.Fatalf`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).IsFailed)
		var ran bool
		sandbox.Run(func() {
			stub.Get(t).Log("-")
			stub.Get(t).Fatalf(`%s %q %s`, `arg1`, `arg2`, `arg3`)
			ran = true
		})
		assert.Must(t).False(ran)
		assert.Must(t).True(stub.Get(t).IsFailed)
		assert.Must(t).Equal(stub.Get(t).Logs.String(), "-\narg1 \"arg2\" arg3\n")
	})

	s.Test(`.Helper`, func(t *testcase.T) {
		stub.Get(t).Helper()
	})

	s.Test(`.Log`, func(t *testcase.T) {
		stb := stub.Get(t)

		stb.Log() // empty log line
		assert.Must(t).Equal("\n", stb.Logs.String())

		stb.Log("foo", "bar", "baz")
		assert.Must(t).Contains(stb.Logs.String(), "\nfoo bar baz\n")

		stb.Log("bar", "baz", "foo")
		assert.Must(t).Contains(stb.Logs.String(), "\nfoo bar baz\nbar baz foo\n")
	})

	s.Test(`.Logf`, func(t *testcase.T) {
		stb := stub.Get(t)

		stb.Logf(`%s %s %q`, `arg1`, `arg2`, `arg3`)
		assert.Must(t).Equal(`arg1 arg2 "arg3"`+"\n", stb.Logs.String())

		stb.Logf(`%s %q %s`, `arg4`, `arg5`, `arg6`)
		assert.Must(t).Equal(`arg1 arg2 "arg3"`+"\n"+`arg4 "arg5" arg6`+"\n", stb.Logs.String())
	})

	s.Context(`.Name`, func(s *testcase.Spec) {
		s.Test(`with provided name, name is used`, func(t *testcase.T) {
			val := t.Random.String()
			stub.Get(t).StubName = val
			assert.Must(t).Equal(val, stub.Get(t).Name())
		})

		s.Test(`without provided name, a name is created and consistently returned`, func(t *testcase.T) {
			stub.Get(t).StubName = ""
			assert.Must(t).NotEmpty(stub.Get(t).Name())
			assert.Must(t).Equal(stub.Get(t).Name(), stub.Get(t).Name())
		})

		s.Test("with provided StubNameFunc", func(t *testcase.T) {
			val := t.Random.String()
			stub.Get(t).StubNameFunc = func() string { return val }
			assert.Must(t).Equal(val, stub.Get(t).Name())
		})
	})

	s.Test(`.Skip`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).Skipped())
		var ran bool
		sandbox.Run(func() {
			stub.Get(t).Skip()
			ran = true
		})
		assert.Must(t).False(ran)
		assert.Must(t).True(stub.Get(t).Skipped())
	})

	s.Test(`.Skip + args`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).Skipped())
		var ran bool
		args := []any{"Hello", "world!"}
		sandbox.Run(func() {
			stub.Get(t).Skip(args...)
			ran = true
		})
		assert.Must(t).False(ran)
		assert.Must(t).True(stub.Get(t).Skipped())
		assert.Must(t).Contains(stub.Get(t).Logs.String(), "Hello world!\n")
	})

	s.Test(`.Skipf + args`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).Skipped())
		var ran bool
		sandbox.Run(func() {
			stub.Get(t).Skipf("%s", "|v|")
			ran = true
		})
		assert.Must(t).False(ran)
		assert.Must(t).True(stub.Get(t).Skipped())
		assert.Must(t).Contains(stub.Get(t).Logs.String(), "|v|\n")
	})

	s.Describe(".SkipNow", func(s *testcase.Spec) {
		s.Test(`goroutine exits`, func(t *testcase.T) {
			var ran bool
			sandbox.Run(func() {
				stub.Get(t).SkipNow()
				ran = true
			})
			assert.Must(t).False(ran)
		})

		s.Test(`.Skipped`, func(t *testcase.T) {
			assert.Must(t).True(!stub.Get(t).Skipped())
			sandbox.Run(func() {
				stub.Get(t).SkipNow()
			})
			assert.Must(t).True(stub.Get(t).Skipped())
		})

		s.Test("OnSkipNow is called", func(t *testcase.T) {
			var fnRan bool
			stub.Get(t).OnSkipNow = func() { fnRan = true }

			sandbox.Run(func() {
				stub.Get(t).SkipNow()
				t.Error("it was not expected to reach this point")
			})
			assert.Must(t).True(stub.Get(t).IsSkipped, "is skipped should be still marked as true")
			assert.Must(t).True(fnRan)
		})
	})

	s.Test(`.Skipf`, func(t *testcase.T) {
		assert.Must(t).True(!stub.Get(t).Skipped())
		var ran bool
		sandbox.Run(func() {
			stub.Get(t).Skipf(`%s`, `arg42`)
			ran = true
		})
		assert.Must(t).False(ran)
		assert.Must(t).True(stub.Get(t).Skipped())
	})

	s.Context(`.TempDir`, func(s *testcase.Spec) {
		s.Test(`with provided temp dir value, value is returned`, func(t *testcase.T) {
			val := t.Random.String()
			stub.Get(t).StubTempDir = val
			assert.Must(t).Equal(val, stub.Get(t).TempDir())
		})
		s.Test(`without a provided temp dir, os temp dir returned`, func(t *testcase.T) {
			assert.Must(t).Equal(os.TempDir(), stub.Get(t).TempDir())
		})
		s.Test("with a testing.TB provided, testing.TB TempDir is created", func(t *testcase.T) {
			st := stub.Get(t)
			st.TB = t
			tmpDir := st.TempDir()

			stat, err := os.Stat(tmpDir)
			assert.Must(t).Nil(err)
			assert.Must(t).True(stat.IsDir())
		})
	})

	s.Test(".Setenv", func(t *testcase.T) {
		key := t.Random.StringNC(5, random.CharsetAlpha())
		t.UnsetEnv(key)
		dtb := stub.Get(t)

		value := t.Random.StringNC(5, random.CharsetAlpha())
		dtb.Setenv(key, value)

		gotValue, hasValue := os.LookupEnv(key)
		assert.Must(t).True(hasValue)
		assert.Must(t).Equal(value, gotValue)

		dtb.Finish()

		_, hasValue = os.LookupEnv(key)
		assert.Must(t).False(hasValue)
	})

	s.Context(".Run", func(s *testcase.Spec) {
		s.Test("the last run's tb can be retrieved", func(t *testcase.T) {
			dtb := &doubles.TB{}

			_, ok := dtb.LastRunTB()
			assert.False(t, ok)

			dtb.Run("", func(tb testing.TB) {})

			ltb, ok := dtb.LastRunTB()
			assert.True(t, ok)
			assert.False(t, ltb.IsFailed)
		})

		s.Test("run tb's name is populated", func(t *testcase.T) {
			dtb := &doubles.TB{}

			dtb.Run("", func(tb testing.TB) {
				assert.NotEmpty(t, strings.TrimPrefix(tb.Name(), dtb.Name()+"/"))
			})

			dtb.Run("name", func(tb testing.TB) {
				assert.Equal(t, dtb.Name()+"/name", tb.Name())
			})

			ltb, ok := dtb.LastRunTB()
			assert.True(t, ok)
			assert.False(t, ltb.IsFailed)
			assert.Equal(t, dtb.Name()+"/name", ltb.Name())
		})

		s.Test("run will encapsulate FailNow goexit", func(t *testcase.T) {
			dtb := &doubles.TB{}
			failNowOut := sandbox.Run(func() {
				var ran bool
				assert.False(t, dtb.Run("", func(tb testing.TB) {
					tb.FailNow()
					ran = true
				}))
				assert.False(t, ran)
			})
			assert.True(t, failNowOut.OK, "fail now should not leak out from the Run block")
			assert.True(t, dtb.IsFailed, "main dtb was expected to fail")

			ltb, ok := dtb.LastRunTB()
			assert.True(t, ok)
			assert.True(t, ltb.IsFailed)
		})

		s.Test("run's tb is failable", func(t *testcase.T) {
			dtb := &doubles.TB{}
			failOut := sandbox.Run(func() {
				var ran bool
				assert.False(t, dtb.Run("", func(tb testing.TB) {
					tb.Fail()
					ran = true
				}))
				assert.True(t, ran)
			})
			assert.True(t, failOut.OK, "fail should not leak out from the Run block")
			assert.True(t, dtb.IsFailed, "main dtb was expected to fail")

			ltb, ok := dtb.LastRunTB()
			assert.True(t, ok)
			assert.True(t, ltb.IsFailed)
		})

	})

	s.Context("#Context", func(s *testcase.Spec) {
		act := let.Act(func(t *testcase.T) context.Context {
			return stub.Get(t).Context()
		})

		s.Test("non nil context returned", func(t *testcase.T) {
			assert.NotNil(t, act(t))
		})

		s.Test("context is cancelled on finish but before cleanup just like testing.TB", func(t *testcase.T) {
			select {
			case <-act(t).Done():
				t.Log("expected that context done is blocking during test")
				t.FailNow()
			default:
			}

			stub.Get(t).Cleanup(func() {
				select {
				case <-act(t).Done():
				default:
					t.Log("expected that context done is not blocking in cleanup")
					t.FailNow()
				}
			})

			stub.Get(t).Finish()
		})
	})

	s.Context("Chdir", func(s *testcase.Spec) {
		s.Test("smoke", func(t *testcase.T) {
			tmpdir := t.TempDir()
			dtb := stub.Get(t)

			og, err := os.Getwd()
			assert.NoError(t, err)
			assert.NotEmpty(t, og)

			assert.NotPanic(t, func() {
				dtb.Chdir(tmpdir)
			})

			cur, err := os.Getwd()
			assert.NoError(t, err)
			assert.True(t, strings.HasSuffix(cur, tmpdir))
			assert.Equal(t, cur, os.Getenv("PWD"))

			assert.NotPanic(t, func() {
				dtb.Finish()
			})

			cur, err = os.Getwd()
			cur = strings.TrimPrefix(cur, "/private")
			assert.NoError(t, err)
			assert.Equal(t, cur, og)
		})
	})

	s.Context("Pass+Passes", func(s *testcase.Spec) {
		s.Test("smoke", func(t *testcase.T) {
			dtb := stub.Get(t)

			passes := t.Random.Repeat(3, 7, func() {
				dtb.Pass()
			})

			assert.Equal(t, passes, dtb.Passes())
		})

		s.Test("race", func(t *testcase.T) {
			dtb := stub.Get(t)

			testcase.Race(func() {
				dtb.Pass()
			}, func() {
				dtb.Pass()
			}, func() {
				dtb.Passes()
			}, func() {
				dtb.Passes()
			})
		})
	})

	s.Context("Timer", func(s *testcase.Spec) {
		s.Test(".IsTimerOn initially returns true (timer on by default)", func(t *testcase.T) {
			dtb := stub.Get(t)
			assert.Must(t).True(dtb.IsTimerOn())
		})

		s.Test(".StopTimer sets timer off (timerOff=0 -> 1)", func(t *testcase.T) {
			dtb := stub.Get(t)
			assert.Must(t).True(dtb.IsTimerOn())
			dtb.StopTimer()
			assert.Must(t).False(dtb.IsTimerOn())
		})

		s.Test(".StartTimer sets timer on (timerOff=1 -> 0)", func(t *testcase.T) {
			dtb := stub.Get(t)
			assert.Must(t).True(dtb.IsTimerOn())
			dtb.StopTimer()
			assert.Must(t).False(dtb.IsTimerOn())
			dtb.StartTimer()
			assert.Must(t).True(dtb.IsTimerOn())
		})

		s.Test(".StartTimer + .StopTimer toggles state", func(t *testcase.T) {
			dtb := stub.Get(t)
			assert.Must(t).True(dtb.IsTimerOn())
			dtb.StopTimer()
			assert.Must(t).False(dtb.IsTimerOn())
			dtb.StartTimer()
			assert.Must(t).True(dtb.IsTimerOn())
			dtb.StopTimer()
			assert.Must(t).False(dtb.IsTimerOn())
			dtb.StartTimer()
			assert.Must(t).True(dtb.IsTimerOn())
		})

		s.Test(".ResetTimer does nothing", func(t *testcase.T) {
			dtb := stub.Get(t)
			assert.Must(t).True(dtb.IsTimerOn())
			dtb.ResetTimer()
			assert.Must(t).True(dtb.IsTimerOn())
			dtb.StopTimer()
			dtb.ResetTimer()
			assert.Must(t).False(dtb.IsTimerOn())
		})

		s.Test("multiple .StartTimer calls (random count), then one .StopTimer -> timer is off", func(t *testcase.T) {
			dtb := stub.Get(t)
			assert.Must(t).True(dtb.IsTimerOn())

			t.Random.Repeat(3, 7, func() {
				dtb.StartTimer()
			})
			// After multiple StartTimer calls, timer should still be on (CAS fails after first)
			assert.Must(t).True(dtb.IsTimerOn())

			dtb.StopTimer()
			assert.Must(t).False(dtb.IsTimerOn())
		})

		s.Test("multiple .StopTimer calls (random count), then one .StartTimer -> timer is on", func(t *testcase.T) {
			dtb := stub.Get(t)
			assert.Must(t).True(dtb.IsTimerOn())

			t.Random.Repeat(3, 7, func() {
				dtb.StopTimer()
			})
			// After multiple StopTimer calls, only first succeeds, timer should be off
			assert.Must(t).False(dtb.IsTimerOn())

			dtb.StartTimer()
			assert.Must(t).True(dtb.IsTimerOn())
		})
	})

	s.Test("doubles.TB implements all testing.TB methods", func(t *testcase.T) {
		var _ testing.TB = (*doubles.TB)(nil)

		var isPromotedMethod = func(tb testing.TB, structType reflect.Type, methodName string) bool {

			method, exists := structType.MethodByName(methodName)
			assert.True(t, exists, assert.MessageF("method %s not found", methodName))
			// Get the function pointer
			funcPtr := method.Func.Pointer()
			// Get runtime function info
			runtimeFunc := runtime.FuncForPC(funcPtr)
			assert.NotNil(t, runtimeFunc, "could not get runtime function info")
			// Check if the source file is "<autogenerated>"
			fileName, _ := runtimeFunc.FileLine(runtimeFunc.Entry())
			// Promoted methods have "<autogenerated>" as their source file
			return fileName == "<autogenerated>"
		}

		type Dummy struct{ testing.TB }

		var (
			TypeTestingTB = reflect.TypeOf((*testing.TB)(nil)).Elem()
			TypeDoublesTB = reflect.TypeOf((**doubles.TB)(nil)).Elem()
			TypeDummy     = reflect.TypeOf((*Dummy)(nil)).Elem()
		)

		assert.True(t, isPromotedMethod(t, TypeDummy, "Helper"))
		assert.True(t, isPromotedMethod(t, TypeDummy, "Cleanup"))
		assert.True(t, isPromotedMethod(t, TypeDummy, "FailNow"))
		assert.True(t, isPromotedMethod(t, TypeDummy, "SkipNow"))

		var TestingTBMethodNumber = TypeTestingTB.NumMethod()
		assert.NotEqual(t, 0, TestingTBMethodNumber)

		for i := 0; i < TestingTBMethodNumber; i++ {
			method := TypeTestingTB.Method(i)
			if !method.IsExported() {
				continue
			}

			name := method.Name

			assert.True(t, isPromotedMethod(t, TypeDummy, name))

			assert.Should(t).False(isPromotedMethod(t, TypeDoublesTB, name),
				assert.MessageF("%s is not defined on *doubles.TB", name))
		}
	})
}

func Test_testingTBContextBehaviour(t *testing.T) {
	assert.NotNil(t, t.Context())

	select {
	case <-t.Context().Done():
		t.Log("expected that context done is blocking during test")
		t.FailNow()
	default:
	}

	t.Cleanup(func() {
		select {
		case <-t.Context().Done():
		default:
			t.Log("expected that context done is not blocking in cleanup")
			t.FailNow()
		}
	})
}

func TestStubTB_implementsTestingTB(t *testing.T) {
	testcase.RunSuite(t, contracts.TestingTB{
		Subject: func(t *testcase.T) testing.TB {
			stb := &doubles.TB{}
			t.Cleanup(stb.Finish)
			return stb
		},
	})
}
