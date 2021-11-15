package internal_test

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/contracts"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/adamluzsi/testcase/internal"
)

var _ testcase.TBRunner = &internal.RecorderTB{}

func TestRecorderTB(t *testing.T) {
	s := testcase.NewSpec(t)

	TB := s.Let(`TB`, func(t *testcase.T) interface{} {
		stub := &internal.StubTB{}
		t.Cleanup(stub.Finish)
		return stub
	})
	tbAsStub := func(t *testcase.T) *internal.StubTB { return TB.Get(t).(*internal.StubTB) }

	recorder := s.Let(`RecorderTB`, func(t *testcase.T) interface{} {
		return &internal.RecorderTB{TB: TB.Get(t).(testing.TB)}
	})
	recorderGet := func(t *testcase.T) *internal.RecorderTB {
		return recorder.Get(t).(*internal.RecorderTB)
	}

	expectToExitGoroutine := func(t *testcase.T, fn func()) {
		_, ok := internal.Recover(func() {
			fn()
		})
		assert.Must(t).False(ok)
	}

	var (
		rndInterfaceListArgs = testcase.Var{
			Name: `args`,
			Init: func(t *testcase.T) interface{} {
				var args []interface{}
				total := fixtures.Random.IntN(12) + 1
				for i := 0; i < total; i++ {
					args = append(args, fixtures.Random.String())
				}
				return args
			},
		}
		rndInterfaceListFormat = testcase.Var{
			Name: `format`,
			Init: func(t *testcase.T) interface{} {
				var format string
				for range rndInterfaceListArgs.Get(t).([]interface{}) {
					format += `%v`
				}
				return format
			},
		}
	)

	thenTBWillMarkedAsFailed := func(s *testcase.Spec, subject func(t *testcase.T)) {
		s.Then(`it will make the TB object failed`, func(t *testcase.T) {
			subject(t)

			assert.Must(t).True(recorderGet(t).IsFailed)
		})
	}

	thenUnderlyingTBWillExpect := func(s *testcase.Spec, subject func(t *testcase.T), fn func(t *testcase.T, stub *internal.StubTB)) {
		s.Then(`on #Forward, the method call is forwarded to the received testing.TB`, func(t *testcase.T) {
			fn(t, tbAsStub(t))
			subject(t)
			internal.Recover(recorderGet(t).Forward)
		})
	}

	s.Test(`by default the TB is not marked as failed`, func(t *testcase.T) {
		assert.Must(t).True(!recorderGet(t).IsFailed)
	})

	s.Describe(`.Fail`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			recorderGet(t).Fail()
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(t *testcase.T, stub *internal.StubTB) {
			t.Cleanup(func() {
				t.Must.True(stub.IsFailed)
			})
		})
	})

	s.Describe(`.FailNow`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, recorderGet(t).FailNow)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(t *testcase.T, stub *internal.StubTB) {
			t.Cleanup(func() {
				t.Must.True(stub.IsFailed)
			})
		})
	})

	s.Describe(`.Error`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			recorderGet(t).Error(`foo`)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(t *testcase.T, stub *internal.StubTB) {
			t.Cleanup(func() {
				t.Must.Contain(stub.Logs, `foo`)
			})
		})
	})

	s.Describe(`.Errorf`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			recorderGet(t).Errorf(`%s -`, `errorf`)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(t *testcase.T, stub *internal.StubTB) {
			t.Cleanup(func() {
				t.Must.Contain(stub.Logs, `errorf -`)
			})
		})
	})

	s.Describe(`.Fatal`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, func() { recorderGet(t).Fatal(`fatal`) })
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(t *testcase.T, stub *internal.StubTB) {
			t.Cleanup(func() {
				t.Must.Contain(stub.Logs, `fatal`)
			})
		})
	})

	s.Describe(`.Fatalf`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, func() { recorderGet(t).Fatalf(`%s -`, `fatalf`) })
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(t *testcase.T, stub *internal.StubTB) {
			t.Cleanup(func() {
				t.Must.Contain(stub.Logs, `fatalf -`)
			})
		})
	})

	s.Describe(`.Failed`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) bool {
			return recorderGet(t).Failed()
		}

		s.When(`failed is`, func(s *testcase.Spec) {
			isFailed := testcase.Var{Name: `failed`}

			s.Before(func(t *testcase.T) {
				recorderGet(t).IsFailed = isFailed.Get(t).(bool)
			})

			s.Context(`true`, func(s *testcase.Spec) {
				isFailed.LetValue(s, true)

				s.Then(`failed will be true`, func(t *testcase.T) {
					assert.Must(t).True(subject(t))
				})

				thenUnderlyingTBWillExpect(s, func(t *testcase.T) { _ = subject(t) }, func(t *testcase.T, stub *internal.StubTB) {
					t.Cleanup(func() {
						t.Must.False(stub.Failed(), "expect that IsFailed don't affect the testing.TB")
					})
				})
			})

			s.Context(`false`, func(s *testcase.Spec) {
				isFailed.LetValue(s, false)

				s.Then(`failed will be false`, func(t *testcase.T) {
					assert.Must(t).False(subject(t))
				})

				thenUnderlyingTBWillExpect(s, func(t *testcase.T) { _ = subject(t) }, func(t *testcase.T, stub *internal.StubTB) {
					t.Cleanup(func() {
						t.Must.False(stub.Failed())
					})
				})
			})
		})
	})

	s.Describe(`.Log`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			recorderGet(t).Log(rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`when no reply is done`, func(t *testcase.T) {
			subject(t)
		})

		s.Test(`on recorder records forward`, func(t *testcase.T) {
			t.Cleanup(func() {
				expected := fmt.Sprintf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
				t.Must.Contain(tbAsStub(t).Logs, expected)
			})
			subject(t)
			recorderGet(t).Forward()
		})
	})

	s.Describe(`.Logf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) {
			recorderGet(t).Logf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`when no reply is done`, func(t *testcase.T) {
			subject(t)
		})

		s.Test(`on recorder records forward`, func(t *testcase.T) {
			t.Cleanup(func() {
				expected := fmt.Sprintf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
				t.Must.Contain(tbAsStub(t).Logs, expected)
			})
			subject(t)
			recorderGet(t).Forward()
		})
	})

	s.Describe(`.Helper`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			recorderGet(t).Helper()
		}

		s.Test(`when no Forward is done`, func(t *testcase.T) {
			subject(t)
		})

		s.Test(`on recorder records forward`, func(t *testcase.T) {
			subject(t)
			recorderGet(t).Forward()
		})
	})

	s.Describe(`.Name`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) string {
			return recorderGet(t).Name()
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			assert.Must(t).Equal(tbAsStub(t).Name(), subject(t))
		})
	})

	s.Describe(`.SkipNow`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			internal.Recover(recorderGet(t).SkipNow)
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			subject(t)
			t.Must.True(tbAsStub(t).IsSkipped)
		})
	})

	s.Describe(`.Skip`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			internal.Recover(func() {
				recorderGet(t).Skip(rndInterfaceListArgs.Get(t).([]interface{})...)
			})
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			t.Cleanup(func() {
				t.Must.True(tbAsStub(t).IsSkipped)
			})
			subject(t)
		})
	})

	s.Describe(`.Skipf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) {
			internal.Recover(func() {
				recorderGet(t).Skipf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
			})
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			t.Cleanup(func() {
				t.Must.True(tbAsStub(t).IsSkipped)
			})
			subject(t)
		})
	})

	s.Describe(`.Skipped`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) bool {
			return recorderGet(t).Skipped()
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			isSkipped := fixtures.Random.Bool()
			tbAsStub(t).IsSkipped = isSkipped
			assert.Must(t).Equal(isSkipped, subject(t))
		})
	})

	s.Describe(`.TempDir`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)

		type TempDirer interface{ TempDir() string }
		var (
			getTempDirer = func(t *testcase.T) TempDirer {
				var rtb interface{} = recorderGet(t)
				td, ok := rtb.(TempDirer)
				if !ok {
					t.Skip(`testing.TB don't support TempDir() string method`)
				}
				return td
			}
			subject = func(t *testcase.T) string {
				return getTempDirer(t).TempDir()
			}
		)

		s.Before(func(t *testcase.T) {
			// early load to force skip for go1.14
			getTempDirer(t)
		})

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			tempDir := fixtures.Random.String()
			tbAsStub(t).StubTempDir = tempDir
			assert.Must(t).Equal(tempDir, subject(t))
		})
	})

	s.Describe(`.Cleanup`, func(s *testcase.Spec) {
		counter := s.LetValue(`cleanup function counter`, 0)
		cleanupFn := s.Let(`cleanup function`, func(t *testcase.T) interface{} {
			return func() { counter.Set(t, counter.Get(t).(int)+1) }
		})
		var subject = func(t *testcase.T) {
			recorderGet(t).Cleanup(cleanupFn.Get(t).(func()))
		}

		s.When(`recorder disposed`, func(s *testcase.Spec) {
			// nothing to do to fulfil this context

			s.Then(`cleanup will never run`, func(t *testcase.T) {
				subject(t)

				assert.Must(t).Equal(0, counter.Get(t))
			})
		})

		s.Test(`when recorder records replied then all event is replied`, func(t *testcase.T) {
			t.Log(`then all records is expected to be replied`)
			stub := tbAsStub(t)
			t.Cleanup(func() {
				t.Must.Contain(stub.Logs, []string{"foo", "bar", "baz"})
			})

			recorderGet(t).Log(`foo`)
			recorderGet(t).Log(`bar`)
			recorderGet(t).Log(`baz`)
			subject(t)
			recorderGet(t).Forward()
			assert.Must(t).Equal(0, counter.Get(t), `Cleanup should not run during reply`)
			stub.Finish()
			assert.Must(t).Equal(1, counter.Get(t), `Cleanup should run on testing.TB finish`)
		})

		s.Test(`on #CleanupNow, only recorder cleanup records should be executed`, func(t *testcase.T) {
			recorderGet(t).Log(`foo`)
			recorderGet(t).Log(`bar`)
			recorderGet(t).Log(`baz`)
			subject(t)

			assert.Must(t).Equal(0, counter.Get(t), `Cleanup should not ran yet`)
			recorderGet(t).CleanupNow()
			assert.Must(t).Equal(1, counter.Get(t), `Cleanup was expected`)
		})

		s.Test(`.Run smoke testing`, func(t *testcase.T) {
			var out []int
			recorderGet(t).Run(``, func(tb testing.TB) {
				tb.Cleanup(func() { out = append(out, 2) })
				tb.Cleanup(func() { out = append(out, 4) })
			})
			assert.Must(t).Equal([]int{4, 2}, out)
		})

		s.When(`goroutine exited because a #FailNow or similar fail function exit the current goroutine`, func(s *testcase.Spec) {
			hasRunFlag := s.LetValue(`has run`, false)
			cleanupFn.Let(s, func(t *testcase.T) interface{} {
				return func() { hasRunFlag.Set(t, true); runtime.Goexit() }
			})

			s.Then(`it should not exit the goroutine that calls #CleanupNow`, func(t *testcase.T) {
				subject(t)
				recorderGet(t).CleanupNow()
				assert.Must(t).True(hasRunFlag.Get(t).(bool))
			})
		})
	})

	s.Describe(`.CleanupNow`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			recorderGet(t).CleanupNow()
		}

		s.When(`passthrough set to`, func(s *testcase.Spec) {
			passthrough := testcase.Var{Name: `passthrough`}
			passthroughGet := func(t *testcase.T) bool { return passthrough.Get(t).(bool) }
			s.Before(func(t *testcase.T) {
				recorderGet(t).Config.Passthrough = passthroughGet(t)
			})

			s.Context(`false`, func(s *testcase.Spec) {
				passthrough.LetValue(s, false)

				s.Then(`config remains unchanged after the play`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).Equal(passthroughGet(t), recorderGet(t).Config.Passthrough)
				})
			})

			s.Context(`true`, func(s *testcase.Spec) {
				passthrough.LetValue(s, true)

				s.Then(`config remains unchanged after the play`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).Equal(passthroughGet(t), recorderGet(t).Config.Passthrough)
				})
			})
		})

		s.When(`no cleanup was called`, func(s *testcase.Spec) {
			s.Then(`it just returns without an issue`, func(t *testcase.T) {
				subject(t)
			})
		})

		s.When(`cleanup has non failing events`, func(s *testcase.Spec) {
			cleanupFootprint := s.Let(`Cleanup Footprint`, func(t *testcase.T) interface{} {
				return []int{}
			})
			cleanupFootprintAppend := func(t *testcase.T, v ...int) {
				cleanupFootprint.Set(t, append(cleanupFootprint.Get(t).([]int), v...))
			}

			s.Before(func(t *testcase.T) {
				recorderGet(t).Cleanup(func() { cleanupFootprintAppend(t, 2) })
				recorderGet(t).Cleanup(func() { cleanupFootprintAppend(t, 4) })
			})

			s.Then(`it will execute cleanups`, func(t *testcase.T) {
				subject(t)

				assert.Must(t).Equal([]int{4, 2}, cleanupFootprint.Get(t))
			})
		})

		s.When(`cleanup has events that fails the test`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				t.Cleanup(func() {
					t.Must.True(tbAsStub(t).IsFailed)
				})
				recorderGet(t).Cleanup(func() { recorderGet(t).FailNow() })
			})

			s.Then(`it will execute cleanups without affecting the current goroutine`, func(t *testcase.T) {
				subject(t)
			})

			s.Then(`it will mark the test failed`, func(t *testcase.T) {
				subject(t)

				assert.Must(t).True(recorderGet(t).IsFailed)
			})
		})

		s.Describe(`idempotent`, func(s *testcase.Spec) {
			s.Test(`calling .CleanupNow multiple times will only replay cleanup once`, func(t *testcase.T) {
				var (
					rtb     = recorderGet(t)
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
					stub    = tbAsStub(t)
					rtb     = recorderGet(t)
					counter int
				)
				stub.StubCleanup = func(f func()) {
					t.Fatal("unexpected .Cleanup call in the testing.TB")
				}
				rtb.Cleanup(func() { counter++ })
				rtb.CleanupNow()
				rtb.Forward()
			})
		})
	})

	s.Describe(`.Forward`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			recorderGet(t).Forward()
			tbAsStub(t).Finish()
		}

		s.When(`.FailNow called in #Cleanup`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				t.Cleanup(func() { t.Must.True(tbAsStub(t).IsFailed) })
				recorderGet(t).Cleanup(func() { recorderGet(t).FailNow() })
			})

			s.Then(`it will replay events to the provided TB`, func(t *testcase.T) {
				subject(t)
			})
		})
	})

	s.Describe(`.Run`, func(s *testcase.Spec) {
		var (
			name    = s.LetValue(`name`, fixtures.Random.String())
			blk     = testcase.Var{Name: `blk`}
			subject = func(t *testcase.T) bool {
				return recorderGet(t).Run(name.Get(t).(string), blk.Get(t).(func(testing.TB)))
			}
		)

		s.When(`block result in a passing sub test`, func(s *testcase.Spec) {
			blk.Let(s, func(t *testcase.T) interface{} {
				return func(testing.TB) {}
			})

			s.Then(`it will report the success`, func(t *testcase.T) {
				assert.Must(t).True(subject(t))
			})

			s.Then(`it will not mark the parent as failed`, func(t *testcase.T) {
				subject(t)

				assert.Must(t).True(!recorderGet(t).IsFailed)
			})
		})

		s.When(`block fails out early`, func(s *testcase.Spec) {
			blk.Let(s, func(t *testcase.T) interface{} {
				return func(tb testing.TB) { tb.FailNow() }
			})

			s.Then(`it will report the markFailed`, func(t *testcase.T) {
				assert.Must(t).True(!subject(t))
			})

			s.Then(`it will mark the parent as failed`, func(t *testcase.T) {
				subject(t)

				assert.Must(t).True(recorderGet(t).IsFailed)
			})
		})
	})
}

func TestRecorderTB_CustomTB_contract(t *testing.T) {
	contracts.CustomTB{
		NewSubject: func(tb testing.TB) testcase.TBRunner {
			stub := &internal.StubTB{}
			rtb := &internal.RecorderTB{TB: stub}
			rtb.Config.Passthrough = true
			return rtb
		},
	}.Test(t)
}

func TestRecorderTB_Record_ConcurrentAccess(t *testing.T) {
	var (
		stub = &internal.StubTB{}
		rtb  = &internal.RecorderTB{TB: stub}
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
