package doubles_test

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"go.llib.dev/testcase/random"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/contracts"
	doubles2 "go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/sandbox"
)

var _ testcase.TBRunner = &doubles2.RecorderTB{}

func TestRecorderTB(t *testing.T) {
	s := testcase.NewSpec(t)

	stubTB := testcase.Let(s, func(t *testcase.T) *doubles2.TB {
		stub := &doubles2.TB{}
		t.Cleanup(stub.Finish)
		return stub
	})
	recorder := testcase.Let(s, func(t *testcase.T) *doubles2.RecorderTB {
		return &doubles2.RecorderTB{TB: stubTB.Get(t)}
	})

	expectToExitGoroutine := func(t *testcase.T, fn func()) {
		out := sandbox.Run(fn)
		assert.Must(t).False(out.OK)
	}

	var (
		rndInterfaceListArgs = testcase.Var[[]any]{
			ID: `args`,
			Init: func(t *testcase.T) []any {
				var args []any
				total := t.Random.IntN(12) + 1
				for i := 0; i < total; i++ {
					args = append(args, t.Random.String())
				}
				return args
			},
		}
		rndInterfaceListFormat = testcase.Var[string]{
			ID: `format`,
			Init: func(t *testcase.T) string {
				var format []string
				for range rndInterfaceListArgs.Get(t) {
					format = append(format, `%v`)
				}
				return strings.Join(format, " ")
			},
		}
	)

	thenTBWillMarkedAsFailed := func(s *testcase.Spec, act func(t *testcase.T)) {
		s.Then(`it will make the TB object failed`, func(t *testcase.T) {
			act(t)

			assert.Must(t).True(recorder.Get(t).IsFailed)
		})
	}

	thenUnderlyingTBWillExpect := func(s *testcase.Spec, subject func(t *testcase.T), fn func(t *testcase.T, stub *doubles2.TB)) {
		s.Then(`on #Forward, the method call is forwarded to the received testing.TB`, func(t *testcase.T) {
			fn(t, stubTB.Get(t))
			subject(t)
			sandbox.Run(recorder.Get(t).Forward)
		})
	}

	s.Test(`by default the TB is not marked as failed`, func(t *testcase.T) {
		assert.Must(t).True(!recorder.Get(t).IsFailed)
	})

	s.Describe(`.Fail`, func(s *testcase.Spec) {
		var act = func(t *testcase.T) {
			recorder.Get(t).Fail()
		}

		thenTBWillMarkedAsFailed(s, act)

		thenUnderlyingTBWillExpect(s, act, func(t *testcase.T, stub *doubles2.TB) {
			t.Cleanup(func() {
				t.Must.True(stub.IsFailed)
			})
		})
	})

	s.Describe(`.FailNow`, func(s *testcase.Spec) {
		var act = func(t *testcase.T) {
			expectToExitGoroutine(t, recorder.Get(t).FailNow)
		}

		thenTBWillMarkedAsFailed(s, act)

		thenUnderlyingTBWillExpect(s, act, func(t *testcase.T, stub *doubles2.TB) {
			t.Cleanup(func() {
				t.Must.True(stub.IsFailed)
			})
		})
	})

	s.Describe(`.Error`, func(s *testcase.Spec) {
		var act = func(t *testcase.T) {
			recorder.Get(t).Error(`foo`)
		}

		thenTBWillMarkedAsFailed(s, act)

		thenUnderlyingTBWillExpect(s, act, func(t *testcase.T, stub *doubles2.TB) {
			t.Cleanup(func() {
				t.Must.Contain(stub.Logs.String(), `foo`)
			})
		})
	})

	s.Describe(`.Errorf`, func(s *testcase.Spec) {
		var act = func(t *testcase.T) {
			recorder.Get(t).Errorf(`%s -`, `errorf`)
		}

		thenTBWillMarkedAsFailed(s, act)

		thenUnderlyingTBWillExpect(s, act, func(t *testcase.T, stub *doubles2.TB) {
			t.Cleanup(func() {
				t.Must.Contain(stub.Logs.String(), `errorf -`)
			})
		})
	})

	s.Describe(`.Fatal`, func(s *testcase.Spec) {
		var act = func(t *testcase.T) {
			expectToExitGoroutine(t, func() { recorder.Get(t).Fatal(`fatal`) })
		}

		thenTBWillMarkedAsFailed(s, act)

		thenUnderlyingTBWillExpect(s, act, func(t *testcase.T, stub *doubles2.TB) {
			t.Cleanup(func() {
				t.Must.Contain(stub.Logs.String(), `fatal`)
			})
		})
	})

	s.Describe(`.Fatalf`, func(s *testcase.Spec) {
		var act = func(t *testcase.T) {
			expectToExitGoroutine(t, func() { recorder.Get(t).Fatalf(`%s -`, `fatalf`) })
		}

		thenTBWillMarkedAsFailed(s, act)

		thenUnderlyingTBWillExpect(s, act, func(t *testcase.T, stub *doubles2.TB) {
			t.Cleanup(func() {
				t.Must.Contain(stub.Logs.String(), `fatalf -`)
			})
		})
	})

	s.Describe(`.Failed`, func(s *testcase.Spec) {
		var act = func(t *testcase.T) bool {
			return recorder.Get(t).Failed()
		}

		s.When(`failed is`, func(s *testcase.Spec) {
			isFailed := testcase.Var[bool]{ID: `failed`}

			s.Before(func(t *testcase.T) {
				recorder.Get(t).IsFailed = isFailed.Get(t)
			})

			s.Context(`true`, func(s *testcase.Spec) {
				isFailed.LetValue(s, true)

				s.Then(`failed will be true`, func(t *testcase.T) {
					assert.Must(t).True(act(t))
				})

				thenUnderlyingTBWillExpect(s, func(t *testcase.T) { _ = act(t) }, func(t *testcase.T, stub *doubles2.TB) {
					t.Cleanup(func() {
						t.Must.False(stub.Failed(), "expect that IsFailed don't affect the testing.TB")
					})
				})
			})

			s.Context(`false`, func(s *testcase.Spec) {
				isFailed.LetValue(s, false)

				s.Then(`failed will be false`, func(t *testcase.T) {
					assert.Must(t).False(act(t))
				})

				thenUnderlyingTBWillExpect(s, func(t *testcase.T) { _ = act(t) }, func(t *testcase.T, stub *doubles2.TB) {
					t.Cleanup(func() {
						t.Must.False(stub.Failed())
					})
				})
			})
		})
	})

	s.Describe(`.Log`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var act = func(t *testcase.T) {
			recorder.Get(t).Log(rndInterfaceListArgs.Get(t)...)
		}

		s.Test(`when no reply is done`, func(t *testcase.T) {
			act(t)
		})

		s.Test(`on recorder records forward`, func(t *testcase.T) {
			t.Cleanup(func() {
				t.Log(rndInterfaceListFormat.Get(t))
				expected := fmt.Sprintf(rndInterfaceListFormat.Get(t)+"\n", rndInterfaceListArgs.Get(t)...)
				t.Must.Contain(stubTB.Get(t).Logs.String(), expected)
			})
			act(t)
			recorder.Get(t).Forward()
		})
	})

	s.Describe(`.Logf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var act = func(t *testcase.T) {
			recorder.Get(t).Logf(rndInterfaceListFormat.Get(t), rndInterfaceListArgs.Get(t)...)
		}

		s.Test(`when no reply is done`, func(t *testcase.T) {
			act(t)
		})

		s.Test(`on recorder records forward`, func(t *testcase.T) {
			t.Cleanup(func() {
				expected := fmt.Sprintf(rndInterfaceListFormat.Get(t), rndInterfaceListArgs.Get(t)...)
				t.Must.Contain(stubTB.Get(t).Logs.String(), expected)
			})
			act(t)
			recorder.Get(t).Forward()
		})
	})

	s.Describe(`.Helper`, func(s *testcase.Spec) {
		var act = func(t *testcase.T) {
			recorder.Get(t).Helper()
		}

		s.Test(`when no Forward is done`, func(t *testcase.T) {
			act(t)
		})

		s.Test(`on recorder records forward`, func(t *testcase.T) {
			act(t)
			recorder.Get(t).Forward()
		})
	})

	s.Describe(`.ID`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) string {
			return recorder.Get(t).Name()
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			assert.Must(t).Equal(stubTB.Get(t).Name(), subject(t))
		})
	})

	s.Describe(`.SkipNow`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			sandbox.Run(recorder.Get(t).SkipNow)
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			subject(t)
			t.Must.True(stubTB.Get(t).IsSkipped)
		})
	})

	s.Describe(`.Skip`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			sandbox.Run(func() {
				recorder.Get(t).Skip(rndInterfaceListArgs.Get(t)...)
			})
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			t.Cleanup(func() {
				t.Must.True(stubTB.Get(t).IsSkipped)
			})
			subject(t)
		})
	})

	s.Describe(`.Skipf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) {
			sandbox.Run(func() {
				recorder.Get(t).Skipf(rndInterfaceListFormat.Get(t), rndInterfaceListArgs.Get(t)...)
			})
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			t.Cleanup(func() {
				t.Must.True(stubTB.Get(t).IsSkipped)
			})
			subject(t)
		})
	})

	s.Describe(`.Skipped`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) bool {
			return recorder.Get(t).Skipped()
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			isSkipped := t.Random.Bool()
			stubTB.Get(t).IsSkipped = isSkipped
			assert.Must(t).Equal(isSkipped, subject(t))
		})
	})

	s.Describe(`.TempDir`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)

		type TempDirer interface{ TempDir() string }
		var (
			getTempDirer = func(t *testcase.T) TempDirer {
				var rtb interface{} = recorder.Get(t)
				td, ok := rtb.(TempDirer)
				if !ok {
					t.Skip(`testing.TB don't support TempDir() string method`)
				}
				return td
			}
		)
		act := func(t *testcase.T) string {
			return getTempDirer(t).TempDir()
		}

		s.Before(func(t *testcase.T) {
			// early load to force skip for go1.14
			getTempDirer(t)
		})

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			tempDir := t.Random.String()
			stubTB.Get(t).StubTempDir = tempDir
			assert.Must(t).Equal(tempDir, act(t))
		})
	})

	s.Describe(`.Cleanup`, func(s *testcase.Spec) {
		counter := testcase.LetValue(s, 0)
		cleanupFn := testcase.Let(s, func(t *testcase.T) func() {
			return func() { counter.Set(t, counter.Get(t)+1) }
		})
		var act = func(t *testcase.T) {
			recorder.Get(t).Cleanup(cleanupFn.Get(t))
		}

		s.When(`recorder disposed`, func(s *testcase.Spec) {
			// nothing to do to fulfil this context

			s.Then(`cleanup will never run`, func(t *testcase.T) {
				act(t)

				assert.Must(t).Equal(0, counter.Get(t))
			})
		})

		s.Test(`when recorder records replied then all event is replied`, func(t *testcase.T) {
			t.Log(`then all records is expected to be replied`)
			stub := stubTB.Get(t)
			t.Cleanup(func() {
				t.Must.Contain(stub.Logs.String(), "foo\nbar\nbaz\n")
			})

			recorder.Get(t).Log(`foo`)
			recorder.Get(t).Log(`bar`)
			recorder.Get(t).Log(`baz`)
			act(t)
			recorder.Get(t).Forward()
			assert.Must(t).Equal(0, counter.Get(t), `Cleanup should not run during reply`)
			stub.Finish()
			assert.Must(t).Equal(1, counter.Get(t), `Cleanup should run on testing.TB finish`)
		})

		s.Test(`on #CleanupNow, only recorder cleanup records should be executed`, func(t *testcase.T) {
			recorder.Get(t).Log(`foo`)
			recorder.Get(t).Log(`bar`)
			recorder.Get(t).Log(`baz`)
			act(t)

			assert.Must(t).Equal(0, counter.Get(t), `Cleanup should not ran yet`)
			recorder.Get(t).CleanupNow()
			assert.Must(t).Equal(1, counter.Get(t), `Cleanup was expected`)
		})

		s.Test(`.Run smoke testing`, func(t *testcase.T) {
			var out []int
			recorder.Get(t).Run(``, func(tb testing.TB) {
				tb.Cleanup(func() { out = append(out, 2) })
				tb.Cleanup(func() { out = append(out, 4) })
			})
			assert.Must(t).Equal([]int{4, 2}, out)
		})

		s.When(`goroutine exited because a #FailNow or similar fail function exit the current goroutine`, func(s *testcase.Spec) {
			hasRunFlag := testcase.LetValue(s, false)
			cleanupFn.Let(s, func(t *testcase.T) func() {
				return func() { hasRunFlag.Set(t, true); runtime.Goexit() }
			})

			s.Then(`it should not exit the goroutine that calls #CleanupNow`, func(t *testcase.T) {
				act(t)
				recorder.Get(t).CleanupNow()
				assert.Must(t).True(hasRunFlag.Get(t))
			})
		})
	})

	s.Describe(`.Setenv`, func(s *testcase.Spec) {
		var (
			key = testcase.Let(s, func(t *testcase.T) string {
				return t.Random.StringNC(t.Random.IntB(5, 10), random.CharsetAlpha())
			})
			value = testcase.Let(s, func(t *testcase.T) string {
				return t.Random.StringNC(t.Random.IntB(5, 10), random.CharsetAlpha())
			})
		)
		var act = func(t *testcase.T) {
			recorder.Get(t).Setenv(key.Get(t), value.Get(t))
		}

		s.Before(func(t *testcase.T) {
			t.UnsetEnv(key.Get(t)) // given the env variable doesn't exists
		})

		s.Test("on use", func(t *testcase.T) {
			act(t)
			env, ok := os.LookupEnv(key.Get(t))
			t.Must.True(ok)
			t.Must.Equal(value.Get(t), env)
		})

		s.Test("on .CleanupNow", func(t *testcase.T) {
			act(t)
			recorder.Get(t).CleanupNow()

			_, ok := os.LookupEnv(key.Get(t))
			t.Must.False(ok)
		})
	})

	s.Describe(`.CleanupNow`, func(s *testcase.Spec) {
		var act = func(t *testcase.T) {
			recorder.Get(t).CleanupNow()
		}

		s.When(`passthrough set to`, func(s *testcase.Spec) {
			passthrough := testcase.Var[bool]{ID: `passthrough`}
			s.Before(func(t *testcase.T) {
				recorder.Get(t).Config.Passthrough = passthrough.Get(t)
			})

			s.Context(`false`, func(s *testcase.Spec) {
				passthrough.LetValue(s, false)

				s.Then(`config remains unchanged after the play`, func(t *testcase.T) {
					act(t)

					assert.Must(t).Equal(passthrough.Get(t), recorder.Get(t).Config.Passthrough)
				})
			})

			s.Context(`true`, func(s *testcase.Spec) {
				passthrough.LetValue(s, true)

				s.Then(`config remains unchanged after the play`, func(t *testcase.T) {
					act(t)

					assert.Must(t).Equal(passthrough.Get(t), recorder.Get(t).Config.Passthrough)
				})
			})
		})

		s.When(`no cleanup was called`, func(s *testcase.Spec) {
			s.Then(`it just returns without an issue`, func(t *testcase.T) {
				act(t)
			})
		})

		s.When(`cleanup has non failing events`, func(s *testcase.Spec) {
			cleanupFootprint := testcase.Let(s, func(t *testcase.T) []int {
				return []int{}
			})
			cleanupFootprintAppend := func(t *testcase.T, v ...int) {
				cleanupFootprint.Set(t, append(cleanupFootprint.Get(t), v...))
			}

			s.Before(func(t *testcase.T) {
				recorder.Get(t).Cleanup(func() { cleanupFootprintAppend(t, 2) })
				recorder.Get(t).Cleanup(func() { cleanupFootprintAppend(t, 4) })
			})

			s.Then(`it will execute cleanups`, func(t *testcase.T) {
				act(t)

				assert.Must(t).Equal([]int{4, 2}, cleanupFootprint.Get(t))
			})
		})

		s.When(`cleanup has events that fails the test`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				t.Cleanup(func() {
					t.Must.True(stubTB.Get(t).IsFailed)
				})
				recorder.Get(t).Cleanup(func() { recorder.Get(t).FailNow() })
			})

			s.Then(`it will execute cleanups without affecting the current goroutine`, func(t *testcase.T) {
				act(t)
			})

			s.Then(`it will mark the test failed`, func(t *testcase.T) {
				act(t)

				assert.Must(t).True(recorder.Get(t).IsFailed)
			})
		})

		s.Describe(`idempotent`, func(s *testcase.Spec) {
			s.Test(`calling .CleanupNow multiple times will only replay cleanup once`, func(t *testcase.T) {
				var (
					rtb     = recorder.Get(t)
					counter int
				)
				rtb.Cleanup(func() { counter++ })
				rtb.Cleanup(func() { counter++ })
				rtb.Cleanup(func() { counter++ })
				//
				rtb.CleanupNow()
				assert.Must(t).Equal(3, counter)
				//
				rtb.CleanupNow()
				assert.Must(t).Equal(3, counter)
			})

			s.Test(`calling .CleanupNow then forward will skip cleanup events`, func(t *testcase.T) {
				var (
					stub    = stubTB.Get(t)
					rtb     = recorder.Get(t)
					counter int
				)
				rtb.Cleanup(func() { counter++ })
				rtb.CleanupNow()
				rtb.Forward()
				stub.Finish() // finish cleanups if there is any
				t.Must.Equal(counter, 1)
			})
		})
	})

	s.Describe(`.Forward`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			recorder.Get(t).Forward()
			stubTB.Get(t).Finish()
		}

		s.When(`.FailNow called in #Cleanup`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				t.Cleanup(func() { t.Must.True(stubTB.Get(t).IsFailed) })
				recorder.Get(t).Cleanup(func() { recorder.Get(t).FailNow() })
			})

			s.Then(`it will replay events to the provided TB`, func(t *testcase.T) {
				subject(t)
			})
		})
	})

	s.Describe(`.Name`, func(s *testcase.Spec) {
		act := func(t *testcase.T) string {
			return recorder.Get(t).Name()
		}

		s.Then("it returns a non-empty name", func(t *testcase.T) {
			assert.NotEmpty(t, act(t))
		})

		s.Then("the name returned is consistent", func(t *testcase.T) {
			assert.Equal(t, act(t), act(t))
		})
	})

	s.Describe(`.Run`, func(s *testcase.Spec) {
		var (
			name = testcase.Let(s, func(t *testcase.T) string {
				return t.Random.String()
			})
			blk = testcase.Var[func(testing.TB)]{ID: `blk`}
			act = func(t *testcase.T) bool {
				return recorder.Get(t).Run(name.Get(t), blk.Get(t))
			}
		)

		s.When(`block result in a passing sub test`, func(s *testcase.Spec) {
			blk.Let(s, func(t *testcase.T) func(testing.TB) {
				return func(testing.TB) {}
			})

			s.Then(`it will report the success`, func(t *testcase.T) {
				assert.Must(t).True(act(t))
			})

			s.Then(`it will not mark the parent as failed`, func(t *testcase.T) {
				act(t)

				assert.Must(t).True(!recorder.Get(t).IsFailed)
			})
		})

		s.When(`block fails out early`, func(s *testcase.Spec) {
			blk.Let(s, func(t *testcase.T) func(testing.TB) {
				return func(tb testing.TB) { tb.FailNow() }
			})

			s.Then(`it will report the markFailed`, func(t *testcase.T) {
				assert.Must(t).True(!act(t))
			})

			s.Then(`it will mark the parent as failed`, func(t *testcase.T) {
				act(t)

				assert.Must(t).True(recorder.Get(t).IsFailed)
			})
		})
	})
}

func TestRecorderTB_implementsCustomTB(t *testing.T) {
	testcase.RunSuite(t, contracts.CustomTB{
		Subject: func(t *testcase.T) testcase.TBRunner {
			stub := &doubles2.TB{}
			t.Defer(stub.Finish)
			rtb := &doubles2.RecorderTB{TB: stub}
			rtb.Config.Passthrough = true
			return rtb
		},
	})
}

func TestRecorderTB_Record_ConcurrentAccess(t *testing.T) {
	var (
		stub = &doubles2.TB{}
		rtb  = &doubles2.RecorderTB{TB: stub}
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		rtb.Log(`first`)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		rtb.Log(`second`)
	}()

	wg.Wait()

	rtb.Forward()
	rtb.CleanupNow()

	wg.Wait()

}
