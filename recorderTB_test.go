package testcase

import (
	"github.com/adamluzsi/testcase/fixtures"
	"sync"
	"testing"

	"github.com/adamluzsi/testcase/internal"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRecorderTB(t *testing.T) {
	s := NewSpec(t)

	recorder := s.Let(`recorderTB`, func(t *T) interface{} {
		return &recorderTB{TB: t.I(`TB`).(testing.TB)}
	})
	getRecorder := func(t *T) *recorderTB {
		return recorder.Get(t).(*recorderTB)
	}

	TB := s.Let(`TB`, func(t *T) interface{} {
		ctrl := gomock.NewController(t)
		t.Defer(ctrl.Finish)
		m := internal.NewMockTB(ctrl)
		return m
	})
	getMockTB := func(t *T) *internal.MockTB { return TB.Get(t).(*internal.MockTB) }

	expectToExitGoroutine := func(t *T, fn func()) {
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

	thenTBWillMarkedAsFailed := func(s *Spec, subject func(t *T)) {
		s.Then(`it will make the TB object failed`, func(t *T) {
			subject(t)

			require.True(t, getRecorder(t).isFailed)
		})
	}

	thenUnderlyingTBWillExpect := func(s *Spec, subject func(t *T), fn func(mock *internal.MockTB)) {
		s.Then(`on #Reply, the method call is replayed to the received testing.TB`, func(t *T) {
			ctrl := gomock.NewController(t)
			t.Defer(ctrl.Finish)
			mockTB := internal.NewMockTB(ctrl)
			fn(mockTB)
			subject(t)
			getRecorder(t).Replay(mockTB)
		})
	}

	s.Test(`by default the TB is not marked as failed`, func(t *T) {
		require.False(t, getRecorder(t).isFailed)
	})

	s.Describe(`#Fail`, func(s *Spec) {
		var subject = func(t *T) {
			getRecorder(t).Fail()
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Fail()
		})
	})

	s.Describe(`#FailNow`, func(s *Spec) {
		var subject = func(t *T) {
			expectToExitGoroutine(t, getRecorder(t).FailNow)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().FailNow()
		})
	})

	s.Describe(`#Error`, func(s *Spec) {
		var subject = func(t *T) {
			getRecorder(t).Error(`foo`)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Error(gomock.Eq(`foo`))
		})
	})

	s.Describe(`#Errorf`, func(s *Spec) {
		var subject = func(t *T) {
			getRecorder(t).Errorf(`%s`, `errorf`)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Errorf(gomock.Eq(`%s`), gomock.Eq(`errorf`))
		})
	})

	s.Describe(`#Fatal`, func(s *Spec) {
		var subject = func(t *T) {
			expectToExitGoroutine(t, func() { getRecorder(t).Fatal(`fatal`) })
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Fatal(gomock.Eq(`fatal`))
		})
	})

	s.Describe(`#Fatalf`, func(s *Spec) {
		var subject = func(t *T) {
			expectToExitGoroutine(t, func() { getRecorder(t).Fatalf(`%s`, `fatalf`) })
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Fatalf(gomock.Eq(`%s`), gomock.Eq(`fatalf`))
		})
	})

	s.Describe(`#Failed`, func(s *Spec) {
		var subject = func(t *T) bool {
			return getRecorder(t).Failed()
		}

		s.Before(func(t *T) {
			getRecorder(t).TB = nil
		})

		s.When(`is failed is`, func(s *Spec) {
			s.Before(func(t *T) {
				getRecorder(t).isFailed = t.I(`failed`).(bool)
			})

			s.Context(`true`, func(s *Spec) {
				s.LetValue(`failed`, true)

				s.Then(`failed will be true`, func(t *T) {
					require.True(t, subject(t))
				})

				thenUnderlyingTBWillExpect(s, func(t *T) { _ = subject(t) }, func(mock *internal.MockTB) {
					mock.EXPECT().Failed()
				})
			})

			s.Context(`false`, func(s *Spec) {
				s.LetValue(`failed`, false)

				s.Then(`failed will be false`, func(t *T) {
					require.False(t, subject(t))
				})

				thenUnderlyingTBWillExpect(s, func(t *T) { _ = subject(t) }, func(mock *internal.MockTB) {
					mock.EXPECT().Failed()
				})

				s.And(`the TB under the hood failed`, func(s *Spec) {
					s.Before(func(t *T) {
						getRecorder(t).TB = &internal.StubTB{IsFailed: true}
					})

					s.Then(`failed will be true`, func(t *T) {
						require.True(t, subject(t))
					})
				})
			})
		})
	})

	s.Describe(`#Log`, func(s *Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *T) {
			getRecorder(t).Log(rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`when no reply is done`, func(t *T) {
			subject(t)
		})

		s.Test(`on recorder events reply`, func(t *T) {
			getMockTB(t).EXPECT().Log(rndInterfaceListArgs.Get(t).([]interface{})...)
			subject(t)
			getRecorder(t).Replay(TB.Get(t).(testing.TB))
		})
	})

	s.Describe(`#Logf`, func(s *Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *T) {
			getRecorder(t).Logf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`when no reply is done`, func(t *T) {
			subject(t)
		})

		s.Test(`on recorder events reply`, func(t *T) {
			getMockTB(t).EXPECT().Logf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
			subject(t)
			getRecorder(t).Replay(TB.Get(t).(testing.TB))
		})
	})

	s.Describe(`#Helper`, func(s *Spec) {
		var subject = func(t *T) {
			getRecorder(t).Helper()
		}

		s.Test(`when no reply is done`, func(t *T) {
			subject(t)
		})

		s.Test(`on recorder events reply`, func(t *T) {
			getMockTB(t).EXPECT().Helper()
			subject(t)
			getRecorder(t).Replay(TB.Get(t).(testing.TB))
		})
	})

	s.Describe(`#Name`, func(s *Spec) {
		var subject = func(t *T) string {
			return getRecorder(t).Name()
		}

		s.Test(`should forward event to parent TB`, func(t *T) {
			name := fixtures.Random.String()
			getMockTB(t).EXPECT().Name().Return(name)
			require.Equal(t, name, subject(t))
		})
	})

	s.Describe(`#SkipNow`, func(s *Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *T) {
			getRecorder(t).SkipNow()
		}

		s.Test(`should forward event to parent TB`, func(t *T) {
			getMockTB(t).EXPECT().SkipNow()
			subject(t)
		})
	})

	s.Describe(`#Skip`, func(s *Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *T) {
			getRecorder(t).Skip(rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`should forward event to parent TB`, func(t *T) {
			getMockTB(t).EXPECT().Skip(rndInterfaceListArgs.Get(t).([]interface{})...)
			subject(t)
		})
	})

	s.Describe(`#Skipf`, func(s *Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *T) {
			getRecorder(t).Skipf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		s.Test(`should forward event to parent TB`, func(t *T) {
			getMockTB(t).EXPECT().Skipf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
			subject(t)
		})
	})

	s.Describe(`#Skipped`, func(s *Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *T) bool {
			return getRecorder(t).Skipped()
		}

		s.Test(`should forward event to parent TB`, func(t *T) {
			isSkipped := fixtures.Random.Bool()
			getMockTB(t).EXPECT().Skipped().Return(isSkipped)
			require.Equal(t, isSkipped, subject(t))
		})
	})

	s.Describe(`#TempDir`, func(s *Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *T) string {
			return getRecorder(t).TempDir()
		}

		s.Test(`should forward event to parent TB`, func(t *T) {
			tempDir := fixtures.Random.String()
			getMockTB(t).EXPECT().TempDir().Return(tempDir)
			require.Equal(t, tempDir, subject(t))
		})
	})

	s.Describe(`#Cleanup`, func(s *Spec) {
		counter := s.LetValue(`cleanup function counter`, 0)
		cleanupFn := s.Let(`cleanup function`, func(t *T) interface{} {
			return func() { counter.Set(t, counter.Get(t).(int)+1) }
		})
		var subject = func(t *T) {
			getRecorder(t).Cleanup(cleanupFn.Get(t).(func()))
		}

		s.When(`recorder disposed`, func(s *Spec) {
			// nothing to do to fulfil this context

			s.Then(`cleanup will never run`, func(t *T) {
				subject(t)

				require.Equal(t, 0, counter.Get(t))
			})
		})

		s.Test(`when recorder events replied then all event is replied`, func(t *T) {
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

		s.Test(`when only recorder cleanup events replied then only cleanup is replied`, func(t *T) {
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

var rndInterfaceListFormat = Var{
	Name: `format`,
	Init: func(t *T) interface{} {
		var format string
		for range rndInterfaceListArgs.Get(t).([]interface{}) {
			format += `%v`
		}
		return format
	},
}

var rndInterfaceListArgs = Var{
	Name: `args`,
	Init: func(t *T) interface{} {
		var args []interface{}
		total := fixtures.Random.IntN(12) + 1
		for i := 0; i < total; i++ {
			args = append(args, fixtures.Random.String())
		}
		return args
	},
}
