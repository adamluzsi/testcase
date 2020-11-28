package internal_test

import (
	"runtime"
	"testing"

	"github.com/adamluzsi/testcase/contracts"
	"github.com/adamluzsi/testcase/internal/mocks"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/adamluzsi/testcase/internal"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var _ testcase.CustomTB = &internal.RecorderTB{}

func TestRecorderTB(t *testing.T) {
	s := testcase.NewSpec(t)

	TB := s.Let(`TB`, func(t *testcase.T) interface{} {
		ctrl := gomock.NewController(t)
		t.Defer(ctrl.Finish)
		m := mocks.NewMockTB(ctrl)
		return m
	})
	tbAsMockGet := func(t *testcase.T) *mocks.MockTB { return TB.Get(t).(*mocks.MockTB) }

	recorder := s.Let(`RecorderTB`, func(t *testcase.T) interface{} {
		return &internal.RecorderTB{TB: TB.Get(t).(testing.TB)}
	})
	recorderGet := func(t *testcase.T) *internal.RecorderTB {
		return recorder.Get(t).(*internal.RecorderTB)
	}

	expectToExitGoroutine := func(t *testcase.T, fn func()) {
		var wasCancelled = true
		internal.InGoroutine(func() {
			fn()
			wasCancelled = false
		})
		require.True(t, wasCancelled)
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

			require.True(t, recorderGet(t).IsFailed)
		})
	}

	thenUnderlyingTBWillExpect := func(s *testcase.Spec, subject func(t *testcase.T), fn func(mock *mocks.MockTB)) {
		s.Then(`on #Forward, the method call is forwarded to the received testing.TB`, func(t *testcase.T) {
			fn(tbAsMockGet(t))
			subject(t)
			recorderGet(t).Forward()
		})
	}

	s.Test(`by default the TB is not marked as failed`, func(t *testcase.T) {
		require.False(t, recorderGet(t).IsFailed)
	})

	s.Describe(`#Fail`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			recorderGet(t).Fail()
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *mocks.MockTB) {
			mock.EXPECT().Fail()
		})
	})

	s.Describe(`#FailNow`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, recorderGet(t).FailNow)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *mocks.MockTB) {
			mock.EXPECT().FailNow()
		})
	})

	s.Describe(`#Error`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			recorderGet(t).Error(`foo`)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *mocks.MockTB) {
			mock.EXPECT().Error(gomock.Eq(`foo`))
		})
	})

	s.Describe(`#Errorf`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			recorderGet(t).Errorf(`%s`, `errorf`)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *mocks.MockTB) {
			mock.EXPECT().Errorf(gomock.Eq(`%s`), gomock.Eq(`errorf`))
		})
	})

	s.Describe(`#Fatal`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, func() { recorderGet(t).Fatal(`fatal`) })
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *mocks.MockTB) {
			mock.EXPECT().Fatal(gomock.Eq(`fatal`))
		})
	})

	s.Describe(`#Fatalf`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, func() { recorderGet(t).Fatalf(`%s`, `fatalf`) })
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *mocks.MockTB) {
			mock.EXPECT().Fatalf(gomock.Eq(`%s`), gomock.Eq(`fatalf`))
		})
	})

	s.Describe(`#Failed`, func(s *testcase.Spec) {
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
					require.True(t, subject(t))
				})

				thenUnderlyingTBWillExpect(s, func(t *testcase.T) { _ = subject(t) }, func(mock *mocks.MockTB) {
					mock.EXPECT().Failed()
				})
			})

			s.Context(`false`, func(s *testcase.Spec) {
				isFailed.LetValue(s, false)

				s.Then(`failed will be false`, func(t *testcase.T) {
					require.False(t, subject(t))
				})

				thenUnderlyingTBWillExpect(s, func(t *testcase.T) { _ = subject(t) }, func(mock *mocks.MockTB) {
					mock.EXPECT().Failed()
				})
			})
		})
	})

	s.Describe(`#Log`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			recorderGet(t).Log(rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`when no reply is done`, func(t *testcase.T) {
			subject(t)
		})

		s.Test(`on recorder records reply`, func(t *testcase.T) {
			tbAsMockGet(t).EXPECT().Log(rndInterfaceListArgs.Get(t).([]interface{})...)
			subject(t)
			recorderGet(t).Forward()
		})
	})

	s.Describe(`#Logf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) {
			recorderGet(t).Logf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`when no reply is done`, func(t *testcase.T) {
			subject(t)
		})

		s.Test(`on recorder records reply`, func(t *testcase.T) {
			tbAsMockGet(t).EXPECT().Logf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
			subject(t)
			recorderGet(t).Forward()
		})
	})

	s.Describe(`#Helper`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			recorderGet(t).Helper()
		}

		s.Test(`when no reply is done`, func(t *testcase.T) {
			subject(t)
		})

		s.Test(`on recorder records reply`, func(t *testcase.T) {
			tbAsMockGet(t).EXPECT().Helper()
			subject(t)
			recorderGet(t).Forward()
		})
	})

	s.Describe(`#Name`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) string {
			return recorderGet(t).Name()
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			name := fixtures.Random.String()
			tbAsMockGet(t).EXPECT().Name().Return(name)
			require.Equal(t, name, subject(t))
		})
	})

	s.Describe(`#SkipNow`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			recorderGet(t).SkipNow()
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			tbAsMockGet(t).EXPECT().SkipNow()
			subject(t)
		})
	})

	s.Describe(`#Skip`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			recorderGet(t).Skip(rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			tbAsMockGet(t).EXPECT().Skip(rndInterfaceListArgs.Get(t).([]interface{})...)
			subject(t)
		})
	})

	s.Describe(`#Skipf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) {
			recorderGet(t).Skipf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			tbAsMockGet(t).EXPECT().Skipf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
			subject(t)
		})
	})

	s.Describe(`#Skipped`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) bool {
			return recorderGet(t).Skipped()
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			isSkipped := fixtures.Random.Bool()
			tbAsMockGet(t).EXPECT().Skipped().Return(isSkipped)
			require.Equal(t, isSkipped, subject(t))
		})
	})

	s.Describe(`#TempDir`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) string {
			return recorderGet(t).TempDir()
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			tempDir := fixtures.Random.String()
			tbAsMockGet(t).EXPECT().TempDir().Return(tempDir)
			require.Equal(t, tempDir, subject(t))
		})
	})

	s.Describe(`#Cleanup`, func(s *testcase.Spec) {
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

				require.Equal(t, 0, counter.Get(t))
			})
		})

		s.Test(`when recorder records replied then all event is replied`, func(t *testcase.T) {
			t.Log(`then all records is expected to be replied`)
			m := tbAsMockGet(t)
			m.EXPECT().Log(gomock.Eq(`foo`))
			m.EXPECT().Log(gomock.Eq(`bar`))
			m.EXPECT().Log(gomock.Eq(`baz`))
			m.EXPECT().Cleanup(gomock.Any()).Do(func(fn func()) { fn() }).AnyTimes()

			recorderGet(t).Log(`foo`)
			recorderGet(t).Log(`bar`)
			recorderGet(t).Log(`baz`)
			subject(t)
			recorderGet(t).Forward()
			require.Equal(t, 1, counter.Get(t), `Cleanup should not run during reply`)
		})

		s.Test(`on #CleanupNow, only recorder cleanup records should be executed`, func(t *testcase.T) {
			recorderGet(t).Log(`foo`)
			recorderGet(t).Log(`bar`)
			recorderGet(t).Log(`baz`)
			subject(t)

			require.Equal(t, 0, counter.Get(t), `Cleanup should not ran yet`)
			recorderGet(t).CleanupNow()
			require.Equal(t, 1, counter.Get(t), `Cleanup was expected`)
		})

		s.Test(`#Run smoke testing`, func(t *testcase.T) {
			var out []int
			recorderGet(t).Run(``, func(tb testing.TB) {
				tb.Cleanup(func() { out = append(out, 2) })
				tb.Cleanup(func() { out = append(out, 4) })
			})
			require.Equal(t, []int{4, 2}, out)
		})

		s.When(`goroutine exited because a #FailNow or similar fail function exit the current goroutine`, func(s *testcase.Spec) {
			hasRunFlag := s.LetValue(`has run`, false)
			cleanupFn.Let(s, func(t *testcase.T) interface{} {
				return func() { hasRunFlag.Set(t, true); runtime.Goexit() }
			})

			s.Then(`it should not exit the goroutine that calls #CleanupNow`, func(t *testcase.T) {
				subject(t)
				recorderGet(t).CleanupNow()
				require.True(t, hasRunFlag.Get(t).(bool))
			})
		})
	})

	s.Describe(`#CleanupNow`, func(s *testcase.Spec) {
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

					require.Equal(t, passthroughGet(t), recorderGet(t).Config.Passthrough)
				})
			})

			s.Context(`true`, func(s *testcase.Spec) {
				passthrough.LetValue(s, true)

				s.Then(`config remains unchanged after the play`, func(t *testcase.T) {
					subject(t)

					require.Equal(t, passthroughGet(t), recorderGet(t).Config.Passthrough)
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

				require.Equal(t, []int{4, 2}, cleanupFootprint.Get(t))
			})
		})

		s.When(`cleanup has events that fails the test`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				tbAsMockGet(t).EXPECT().FailNow()
				recorderGet(t).Cleanup(func() { recorderGet(t).FailNow() })
			})

			s.Then(`it will execute cleanups without affecting the current goroutine`, func(t *testcase.T) {
				subject(t)
			})

			s.Then(`it will mark the test failed`, func(t *testcase.T) {
				subject(t)

				require.True(t, recorderGet(t).IsFailed)
			})
		})
	})

	s.Describe(`#Forward`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			recorderGet(t).Forward()
		}

		s.When(`#FailNow called in #Cleanup`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				tbAsMockGet(t).EXPECT().Cleanup(gomock.Any()).Do(func(f func()) { f() })
				tbAsMockGet(t).EXPECT().FailNow()
				recorderGet(t).Cleanup(func() { recorderGet(t).FailNow() })
			})

			s.Then(`it will replay events to the provided TB`, func(t *testcase.T) {
				subject(t)
			})
		})
	})

	s.Describe(`#Run`, func(s *testcase.Spec) {
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
				require.True(t, subject(t))
			})

			s.Then(`it will not mark the parent as failed`, func(t *testcase.T) {
				subject(t)

				require.False(t, recorderGet(t).IsFailed)
			})
		})

		s.When(`block fails out early`, func(s *testcase.Spec) {
			blk.Let(s, func(t *testcase.T) interface{} {
				return func(tb testing.TB) { tb.FailNow() }
			})

			s.Then(`it will report the markFailed`, func(t *testcase.T) {
				require.False(t, subject(t))
			})

			s.Then(`it will mark the parent as failed`, func(t *testcase.T) {
				subject(t)

				require.True(t, recorderGet(t).IsFailed)
			})
		})
	})
}

func TestRecorderTB_CustomTB_contract(t *testing.T) {
	contracts.CustomTB{
		NewSubject: func(tb testing.TB) testcase.CustomTB {
			mock := mocks.NewWithDefaults(tb, func(*mocks.MockTB) {})
			rtb := &internal.RecorderTB{TB: mock}
			rtb.Config.Passthrough = true
			return rtb
		},
	}.Test(t)
}
