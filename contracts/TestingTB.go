package contracts

import (
	"os"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal"
	"github.com/adamluzsi/testcase/random"
)

type TestingTB struct {
	Subject func(*testcase.T) testing.TB
}

func (c TestingTB) Spec(s *testcase.Spec) {
	testingTB := testcase.Let(s, func(t *testcase.T) testing.TB {
		return c.Subject(t)
	})

	expectToExitGoroutine := func(t *testcase.T, fn func()) {
		_, ok := internal.Recover(fn)
		t.Must.False(ok)
	}

	var (
		rnd                  = random.New(random.CryptoSeed{})
		rndInterfaceListArgs = testcase.Var[[]any]{
			ID: `args`,
			Init: func(t *testcase.T) []any {
				var args []any
				total := rnd.IntN(12) + 1
				for i := 0; i < total; i++ {
					args = append(args, rnd.String())
				}
				return args
			},
		}
		rndInterfaceListFormat = testcase.Var[string]{
			ID: `format`,
			Init: func(t *testcase.T) string {
				var format string
				for range rndInterfaceListArgs.Get(t) {
					format += `%v`
				}
				return format
			},
		}
	)

	thenItWillMarkTheTestAsFailed := func(s *testcase.Spec, subject func(t *testcase.T)) {
		s.Then(`it will mark the test as failed`, func(t *testcase.T) {
			subject(t)

			assert.Must(t).True(testingTB.Get(t).Failed())
		})
	}

	thenItWillNotMarkTheTestAsFailed := func(s *testcase.Spec, subject func(t *testcase.T)) {
		s.Then(`it will not mark the test as failed`, func(t *testcase.T) {
			subject(t)

			assert.Must(t).True(!testingTB.Get(t).Failed())
		})
	}

	s.Describe(`#Fail`, func(s *testcase.Spec) {
		act := func(t *testcase.T) {
			testingTB.Get(t).Fail()
		}

		thenItWillMarkTheTestAsFailed(s, act)
	})

	s.Describe(`#FailNow`, func(s *testcase.Spec) {
		act := func(t *testcase.T) {
			expectToExitGoroutine(t, testingTB.Get(t).FailNow)
		}

		thenItWillMarkTheTestAsFailed(s, act)
	})

	s.Describe(`#Error`, func(s *testcase.Spec) {
		act := func(t *testcase.T) {
			testingTB.Get(t).Error(`foo`)
		}

		thenItWillMarkTheTestAsFailed(s, act)
	})

	s.Describe(`#Errorf`, func(s *testcase.Spec) {
		act := func(t *testcase.T) {
			testingTB.Get(t).Errorf(`%s %s %s`, `foo`, `bar`, `baz`)
		}

		thenItWillMarkTheTestAsFailed(s, act)
	})

	s.Describe(`#Fatal`, func(s *testcase.Spec) {
		act := func(t *testcase.T) {
			expectToExitGoroutine(t, func() { testingTB.Get(t).Fatal(`fatal`) })
		}

		thenItWillMarkTheTestAsFailed(s, act)
	})

	s.Describe(`#Fatalf`, func(s *testcase.Spec) {
		act := func(t *testcase.T) {
			expectToExitGoroutine(t, func() { testingTB.Get(t).Fatalf(`%s`, `fatalf`) })
		}

		thenItWillMarkTheTestAsFailed(s, act)
	})

	s.Describe(`#Failed`, func(s *testcase.Spec) {
		act := func(t *testcase.T) bool {
			return testingTB.Get(t).Failed()
		}

		thenItWillNotMarkTheTestAsFailed(s, func(t *testcase.T) { act(t) })

		s.Test(`when test is green by default, it returns false`, func(t *testcase.T) {
			assert.Must(t).True(!act(t))
		})

		s.When(`test failed`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				testingTB.Get(t).Fail()
			})

			s.Then(`it will be true`, func(t *testcase.T) {
				assert.Must(t).True(act(t))
			})
		})
	})

	s.Describe(`#Log`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		act := func(t *testcase.T) {
			testingTB.Get(t).Log(rndInterfaceListArgs.Get(t)...)
		}

		thenItWillNotMarkTheTestAsFailed(s, act)
	})

	s.Describe(`#Logf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		act := func(t *testcase.T) {
			testingTB.Get(t).Logf(rndInterfaceListFormat.Get(t), rndInterfaceListArgs.Get(t)...)
		}

		thenItWillNotMarkTheTestAsFailed(s, act)
	})

	s.Describe(`#Helper`, func(s *testcase.Spec) {
		act := func(t *testcase.T) {
			testingTB.Get(t).Helper()
		}

		thenItWillNotMarkTheTestAsFailed(s, act)
	})

	s.Describe(`#Name`, func(s *testcase.Spec) {
		act := func(t *testcase.T) string {
			return testingTB.Get(t).Name()
		}

		thenItWillNotMarkTheTestAsFailed(s, func(t *testcase.T) { act(t) })

		s.Then(`the name returned is not empty`, func(t *testcase.T) {
			t.Must.NotEqual(0, len(act(t)))
		})
	})

	s.Describe(`#SkipNow`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		act := func(t *testcase.T) {
			expectToExitGoroutine(t, testingTB.Get(t).SkipNow)
		}

		thenItWillNotMarkTheTestAsFailed(s, act)
	})

	s.Describe(`#Skip`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		act := func(t *testcase.T) {
			expectToExitGoroutine(t, func() {
				testingTB.Get(t).Skip(rndInterfaceListArgs.Get(t)...)
			})
		}

		thenItWillNotMarkTheTestAsFailed(s, act)
	})

	s.Describe(`#Skipf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		act := func(t *testcase.T) {
			expectToExitGoroutine(t, func() {
				testingTB.Get(t).Skipf(rndInterfaceListFormat.Get(t), rndInterfaceListArgs.Get(t)...)
			})
		}

		thenItWillNotMarkTheTestAsFailed(s, act)
	})

	s.Describe(`#Skipped`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		act := func(t *testcase.T) bool {
			return testingTB.Get(t).Skipped()
		}

		thenItWillNotMarkTheTestAsFailed(s, func(t *testcase.T) { act(t) })

		s.When(`test was skipped`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				expectToExitGoroutine(t, testingTB.Get(t).SkipNow)
			})

			s.Then(`it will report that test was skipped`, func(t *testcase.T) {
				assert.Must(t).True(act(t))
			})
		})

		s.Test(`by default tests are not skipped so it will report false`, func(t *testcase.T) {
			assert.Must(t).True(!act(t))
		})
	})

	s.Describe(`#TempDir`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)

		type TempDirer interface{ TempDir() string }
		var (
			getTempDirer = func(t *testcase.T) TempDirer {
				td, ok := testingTB.Get(t).(TempDirer)
				if !ok {
					t.Skip(`testing.TB don't support TempDir() string method`)
				}
				return td
			}
			act = func(t *testcase.T) string {
				return getTempDirer(t).TempDir()
			}
		)

		thenItWillNotMarkTheTestAsFailed(s, func(t *testcase.T) { act(t) })

		s.Then(`return an existing directory`, func(t *testcase.T) {
			tmpdir := act(t)

			t.Must.True(0 < len(tmpdir))
			if fi, err := os.Stat(tmpdir); err != nil {
				assert.Must(t).True(!os.IsNotExist(err), `expected to !os.IsNotExist`)
			} else {
				assert.Must(t).True(fi.Mode().IsDir())
			}
		})
	})

	s.Describe(`#Cleanup`, func(s *testcase.Spec) {
		s.HasSideEffect()
		var cleanups []int
		s.AfterAll(func(tb testing.TB) {
			assert.Equal(tb, []int{4, 2}, cleanups)
		})
		s.Test(``, func(t *testcase.T) {
			testingTB.Get(t).Cleanup(func() { cleanups = append(cleanups, 2) })
			testingTB.Get(t).Cleanup(func() { cleanups = append(cleanups, 4) })
		})
	})
}
