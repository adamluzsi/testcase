package testcase

import (
	"sync"
	"testing"

	"github.com/adamluzsi/testcase/internal"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRecorderTB(t *testing.T) {
	s := NewSpec(t)

	s.Let(recorderTBVN, func(t *T) interface{} {
		return &recorderTB{TB: t.I(`TB`).(testing.TB)}
	})

	s.Let(`TB`, func(t *T) interface{} {
		return &testing.T{}
	})

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

			require.True(t, getRecorderTB(t).isFailed)
		})
	}

	thenUnderlyingTBWillExpect := func(s *Spec, subject func(t *T), fn func(mock *internal.MockTB)) {
		s.Then(`on #Reply, the method call is replayed to the received testing.TB`, func(t *T) {
			ctrl := gomock.NewController(t)
			t.Defer(ctrl.Finish)
			mockTB := internal.NewMockTB(ctrl)
			fn(mockTB)
			subject(t)
			getRecorderTB(t).Replay(mockTB)
		})
	}

	s.Test(`by default the TB is not marked as failed`, func(t *T) {
		require.False(t, getRecorderTB(t).isFailed)
	})

	s.Describe(`#Fail`, func(s *Spec) {
		var subject = func(t *T) {
			getRecorderTB(t).Fail()
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Fail()
		})
	})

	s.Describe(`#FailNow`, func(s *Spec) {
		var subject = func(t *T) {
			expectToExitGoroutine(t, getRecorderTB(t).FailNow)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().FailNow()
		})
	})

	s.Describe(`#Error`, func(s *Spec) {
		var subject = func(t *T) {
			getRecorderTB(t).Error(`foo`)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Error(gomock.Eq(`foo`))
		})
	})

	s.Describe(`#Errorf`, func(s *Spec) {
		var subject = func(t *T) {
			getRecorderTB(t).Errorf(`%s`, `errorf`)
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Errorf(gomock.Eq(`%s`), gomock.Eq(`errorf`))
		})
	})

	s.Describe(`#Fatal`, func(s *Spec) {
		var subject = func(t *T) {
			expectToExitGoroutine(t, func() { getRecorderTB(t).Fatal(`fatal`) })
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Fatal(gomock.Eq(`fatal`))
		})
	})

	s.Describe(`#Fatalf`, func(s *Spec) {
		var subject = func(t *T) {
			expectToExitGoroutine(t, func() { getRecorderTB(t).Fatalf(`%s`, `fatalf`) })
		}

		thenTBWillMarkedAsFailed(s, subject)

		thenUnderlyingTBWillExpect(s, subject, func(mock *internal.MockTB) {
			mock.EXPECT().Fatalf(gomock.Eq(`%s`), gomock.Eq(`fatalf`))
		})
	})

	s.Describe(`#Failed`, func(s *Spec) {
		var subject = func(t *T) bool {
			return getRecorderTB(t).Failed()
		}

		s.Before(func(t *T) {
			getRecorderTB(t).TB = nil
		})

		s.When(`is failed is`, func(s *Spec) {
			s.Before(func(t *T) {
				getRecorderTB(t).isFailed = t.I(`failed`).(bool)
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
						getRecorderTB(t).TB = &internal.StubTB{IsFailed: true}
					})

					s.Then(`failed will be true`, func(t *T) {
						require.True(t, subject(t))
					})
				})
			})
		})
	})
}

const recorderTBVN = `Non failing TB`

func getRecorderTB(t *T) *recorderTB {
	return t.I(recorderTBVN).(*recorderTB)
}
