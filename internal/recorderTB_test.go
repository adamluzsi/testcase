package internal_test

import (
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/adamluzsi/testcase/internal"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRecorderTB(t *testing.T) {
	s := testcase.NewSpec(t)

	recorder := s.Let(`RecorderTB`, func(t *testcase.T) interface{} {
		return &internal.RecorderTB{TB: t.I(`TB`).(testing.TB)}
	})
	getRecorder := func(t *testcase.T) *internal.RecorderTB {
		return recorder.Get(t).(*internal.RecorderTB)
	}

	TB := s.Let(`TB`, func(t *testcase.T) interface{} {
		ctrl := gomock.NewController(t)
		t.Defer(ctrl.Finish)
		m := internal.NewMockTB(ctrl)
		return m
	})
	getMockTB := func(t *testcase.T) *internal.MockTB { return TB.Get(t).(*internal.MockTB) }

	expectToExitGoroutine := func(t *testcase.T, fn func()) {
		var wg sync.WaitGroup
		wg.Add(1)
		var wasCancelled = true
		go func() {
			defer wg.Done()
			fn()
			wasCancelled = false
		}()
		wg.Wait()
		require.True(t, wasCancelled)
	}

	thenTBWillMarkedAsFailed := func(s *testcase.Spec, subject func(t *testcase.T)) {
		s.Then(`it will make the TB object failed`, func(t *testcase.T) {
			subject(t)

			require.True(t, getRecorder(t).IsFailed)
		})
	}

	thenUnderlyingTBWillExpect := func(s *testcase.Spec, subject func(t *testcase.T), fn func(mock *internal.MockTB)) {
		s.Then(`on #Reply, the method call is replayed to the received testing.TB`, func(t *testcase.T) {
			ctrl := gomock.NewController(t)
			t.Defer(ctrl.Finish)
			mockTB := internal.NewMockTB(ctrl)
			fn(mockTB)
			subject(t)
			getRecorder(t).Replay(mockTB)
		})
	}

	s.Test(`by default the TB is not marked as failed`, func(t *testcase.T) {
		require.False(t, getRecorder(t).IsFailed)
	})

	s.Describe(`#Fail`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			getRecorder(t).Fail()
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Fail()
		})
	})

	s.Describe(`#FailNow`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, getRecorder(t).FailNow)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().FailNow()
		})
	})

	s.Describe(`#Error`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			getRecorder(t).Error(`foo`)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Error(gomock.Eq(`foo`))
		})
	})

	s.Describe(`#Errorf`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			getRecorder(t).Errorf(`%s`, `errorf`)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Errorf(gomock.Eq(`%s`), gomock.Eq(`errorf`))
		})
	})

	s.Describe(`#Fatal`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, func() { getRecorder(t).Fatal(`fatal`) })
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Fatal(gomock.Eq(`fatal`))
		})
	})

	s.Describe(`#Fatalf`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, func() { getRecorder(t).Fatalf(`%s`, `fatalf`) })
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Fatalf(gomock.Eq(`%s`), gomock.Eq(`fatalf`))
		})
	})

	s.Describe(`#Failed`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) bool {
			return getRecorder(t).Failed()
		}

		s.Before(func(t *testcase.T) {
			getRecorder(t).TB = nil
		})

		s.When(`is failed is`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				getRecorder(t).IsFailed = t.I(`failed`).(bool)
			})

			s.Context(`true`, func(s *testcase.Spec) {
				s.LetValue(`failed`, true)

				s.Then(`failed will be true`, func(t *testcase.T) {
					require.True(t, subject(t))
				})

				thenUnderlyingTBWillExpect(s, func(t *testcase.T) { _ = subject(t) }, func(mock *internal.MockTB) {
					mock.EXPECT().Failed()
				})
			})

			s.Context(`false`, func(s *testcase.Spec) {
				s.LetValue(`failed`, false)

				s.Then(`failed will be false`, func(t *testcase.T) {
					require.False(t, subject(t))
				})

				thenUnderlyingTBWillExpect(s, func(t *testcase.T) { _ = subject(t) }, func(mock *internal.MockTB) {
					mock.EXPECT().Failed()
				})

				s.And(`the TB under the hood failed`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						getRecorder(t).TB = &internal.StubTB{IsFailed: true}
					})

					s.Then(`failed will be true`, func(t *testcase.T) {
						require.True(t, subject(t))
					})
				})
			})
		})
	})

	s.Describe(`#Log`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			getRecorder(t).Log(rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`when no reply is done`, func(t *testcase.T) {
			subject(t)
		})

		s.Test(`on recorder events reply`, func(t *testcase.T) {
			getMockTB(t).EXPECT().Log(rndInterfaceListArgs.Get(t).([]interface{})...)
			subject(t)
			getRecorder(t).Replay(TB.Get(t).(testing.TB))
		})
	})

	s.Describe(`#Logf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) {
			getRecorder(t).Logf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`when no reply is done`, func(t *testcase.T) {
			subject(t)
		})

		s.Test(`on recorder events reply`, func(t *testcase.T) {
			getMockTB(t).EXPECT().Logf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
			subject(t)
			getRecorder(t).Replay(TB.Get(t).(testing.TB))
		})
	})

	s.Describe(`#Helper`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			getRecorder(t).Helper()
		}

		s.Test(`when no reply is done`, func(t *testcase.T) {
			subject(t)
		})

		s.Test(`on recorder events reply`, func(t *testcase.T) {
			getMockTB(t).EXPECT().Helper()
			subject(t)
			getRecorder(t).Replay(TB.Get(t).(testing.TB))
		})
	})

	s.Describe(`#Name`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) string {
			return getRecorder(t).Name()
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			name := fixtures.Random.String()
			getMockTB(t).EXPECT().Name().Return(name)
			require.Equal(t, name, subject(t))
		})
	})

	s.Describe(`#SkipNow`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			getRecorder(t).SkipNow()
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			getMockTB(t).EXPECT().SkipNow()
			subject(t)
		})
	})

	s.Describe(`#Skip`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			getRecorder(t).Skip(rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			getMockTB(t).EXPECT().Skip(rndInterfaceListArgs.Get(t).([]interface{})...)
			subject(t)
		})
	})

	s.Describe(`#Skipf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) {
			getRecorder(t).Skipf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			getMockTB(t).EXPECT().Skipf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
			subject(t)
		})
	})

	s.Describe(`#Skipped`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) bool {
			return getRecorder(t).Skipped()
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			isSkipped := fixtures.Random.Bool()
			getMockTB(t).EXPECT().Skipped().Return(isSkipped)
			require.Equal(t, isSkipped, subject(t))
		})
	})

	s.Describe(`#TempDir`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) string {
			return getRecorder(t).TempDir()
		}

		s.Test(`should forward event to parent TB`, func(t *testcase.T) {
			tempDir := fixtures.Random.String()
			getMockTB(t).EXPECT().TempDir().Return(tempDir)
			require.Equal(t, tempDir, subject(t))
		})
	})

	s.Describe(`#Cleanup`, func(s *testcase.Spec) {
		counter := s.LetValue(`cleanup function counter`, 0)
		cleanupFn := s.Let(`cleanup function`, func(t *testcase.T) interface{} {
			return func() { counter.Set(t, counter.Get(t).(int)+1) }
		})
		var subject = func(t *testcase.T) {
			getRecorder(t).Cleanup(cleanupFn.Get(t).(func()))
		}

		s.When(`recorder disposed`, func(s *testcase.Spec) {
			// nothing to do to fulfil this context

			s.Then(`cleanup will never run`, func(t *testcase.T) {
				subject(t)

				require.Equal(t, 0, counter.Get(t))
			})
		})

		s.Test(`when recorder events replied then all event is replied`, func(t *testcase.T) {
			t.Log(`then all events is expected to be replied`)
			m := getMockTB(t)
			m.EXPECT().Log(gomock.Eq(`foo`))
			m.EXPECT().Log(gomock.Eq(`bar`))
			m.EXPECT().Log(gomock.Eq(`baz`))
			m.EXPECT().Cleanup(gomock.Any()).Do(func(fn func()) { fn() })

			getRecorder(t).Log(`foo`)
			getRecorder(t).Log(`bar`)
			getRecorder(t).Log(`baz`)
			subject(t)
			getRecorder(t).Replay(TB.Get(t).(testing.TB))
			require.Equal(t, 1, counter.Get(t))
		})

		s.Test(`when only recorder cleanup events replied then only cleanup is replied`, func(t *testcase.T) {
			t.Log(`only cleanup is expected in the reply`)
			getMockTB(t).EXPECT().Cleanup(gomock.Any()).Do(func(fn func()) { fn() })

			getRecorder(t).Log(`foo`)
			getRecorder(t).Log(`bar`)
			getRecorder(t).Log(`baz`)
			subject(t)

			getRecorder(t).ReplayCleanup(TB.Get(t).(testing.TB))
			require.Equal(t, 1, counter.Get(t))
		})

		//s.Test(`when reply for everything but cleanup events requested`, func(t *T) {
		//	t.Log(`cleanup is not expected in the reply`)
		//	m := getMockTB(t)
		//	m.EXPECT().Log(gomock.Eq(`foo`))
		//	m.EXPECT().Log(gomock.Eq(`bar`))
		//	m.EXPECT().Log(gomock.Eq(`baz`))
		//
		//	getRecorder(t).Log(`foo`)
		//	getRecorder(t).Log(`bar`)
		//	getRecorder(t).Log(`baz`)
		//	subject(t)
		//
		//	getRecorder(t).ReplayWithoutCleanup(TB.Get(t).(testing.TB))
		//
		//	require.Equal(t, 0, counter.Get(t))
		//})
	})
}

var rndInterfaceListFormat = testcase.Var{
	Name: `format`,
	Init: func(t *testcase.T) interface{} {
		var format string
		for range rndInterfaceListArgs.Get(t).([]interface{}) {
			format += `%v`
		}
		return format
	},
}

var rndInterfaceListArgs = testcase.Var{
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
