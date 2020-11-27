package contracts

import (
	"os"
	"sync"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"
)

type CustomTB struct {
	NewSubject func(testing.TB) testcase.CustomTB
}

func (spec CustomTB) Test(t *testing.T) { spec.Spec(t) }

func (spec CustomTB) Benchmark(b *testing.B) { spec.Spec(b) }

func (spec CustomTB) Spec(tb testing.TB) {
	s := testcase.NewSpec(tb)

	customTB := s.Let(`Custom TB implementation`, func(t *testcase.T) interface{} {
		return spec.NewSubject(t)
	})
	customTBGet := func(t *testcase.T) testcase.CustomTB {
		return customTB.Get(t).(testcase.CustomTB)
	}

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

	thenItWillMarkTheTestAsFailed := func(s *testcase.Spec, subject func(t *testcase.T)) {
		s.Then(`it will mark the test as failed`, func(t *testcase.T) {
			subject(t)

			require.True(t, customTBGet(t).Failed())
		})
	}

	thenItWillNotMarkTheTestAsFailed := func(s *testcase.Spec, subject func(t *testcase.T)) {
		s.Then(`it will not mark the test as failed`, func(t *testcase.T) {
			subject(t)

			require.False(t, customTBGet(t).Failed())
		})
	}

	s.Describe(`#Fail`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			customTBGet(t).Fail()
		}

		thenItWillMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#FailNow`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, customTBGet(t).FailNow)
		}

		thenItWillMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Error`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			customTBGet(t).Error(`foo`)
		}

		thenItWillMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Errorf`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			customTBGet(t).Errorf(`%s %s %s`, `foo`, `bar`, `baz`)
		}

		thenItWillMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Fatal`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, func() { customTBGet(t).Fatal(`fatal`) })
		}

		thenItWillMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Fatalf`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, func() { customTBGet(t).Fatalf(`%s`, `fatalf`) })
		}

		thenItWillMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Failed`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) bool {
			return customTBGet(t).Failed()
		}

		thenItWillNotMarkTheTestAsFailed(s, func(t *testcase.T) { subject(t) })

		s.Test(`when test is green by default, it returns false`, func(t *testcase.T) {
			require.False(t, subject(t))
		})

		s.When(`test failed`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				customTBGet(t).Fail()
			})

			s.Then(`it will be true`, func(t *testcase.T) {
				require.True(t, subject(t))
			})
		})
	})

	s.Describe(`#Log`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			customTBGet(t).Log(rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		thenItWillNotMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Logf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) {
			customTBGet(t).Logf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
		}

		thenItWillNotMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Helper`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			customTBGet(t).Helper()
		}

		thenItWillNotMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Name`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) string {
			return customTBGet(t).Name()
		}

		thenItWillNotMarkTheTestAsFailed(s, func(t *testcase.T) { subject(t) })

		s.Then(`the name returned is not empty`, func(t *testcase.T) {
			require.NotEmpty(t, subject(t))
		})
	})

	s.Describe(`#SkipNow`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, customTBGet(t).SkipNow)
		}

		thenItWillNotMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Skip`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, func() {
				customTBGet(t).Skip(rndInterfaceListArgs.Get(t).([]interface{})...)
			})
		}

		thenItWillNotMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Skipf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, func() {
				customTBGet(t).Skipf(rndInterfaceListFormat.Get(t).(string), rndInterfaceListArgs.Get(t).([]interface{})...)
			})
		}

		thenItWillNotMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Skipped`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) bool {
			return customTBGet(t).Skipped()
		}

		thenItWillNotMarkTheTestAsFailed(s, func(t *testcase.T) { subject(t) })

		s.When(`test was skipped`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				expectToExitGoroutine(t, customTBGet(t).SkipNow)
			})

			s.Then(`it will report that test was skipped`, func(t *testcase.T) {
				require.True(t, subject(t))
			})
		})

		s.Test(`by default tests are not skipped so it will report false`, func(t *testcase.T) {
			require.False(t, subject(t))
		})
	})

	s.Describe(`#TempDir`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) string {
			return customTBGet(t).TempDir()
		}

		thenItWillNotMarkTheTestAsFailed(s, func(t *testcase.T) { subject(t) })

		s.Then(`return an existing directory`, func(t *testcase.T) {
			tmpdir := subject(t)

			require.NotEmpty(t, tmpdir)
			if fi, err := os.Stat(tmpdir); err != nil {
				require.False(t, os.IsNotExist(err), `expected to !os.IsNotExist`)
			} else {
				require.True(t, fi.Mode().IsDir())
			}
		})
	})

	s.Describe(`#Cleanup`, func(s *testcase.Spec) {
		s.Test(`smoke testing`, func(t *testcase.T) {
			var cleanups []int
			customTBGet(t).Run(``, func(tb testing.TB) {
				tb.Cleanup(func() { cleanups = append(cleanups, 2) })
				tb.Cleanup(func() { cleanups = append(cleanups, 4) })
			})

			require.Equal(t, []int{4, 2}, cleanups)
		})
	})

	s.Describe(`#Run`, func(s *testcase.Spec) {
		var (
			name    = s.LetValue(`name`, fixtures.Random.String())
			blk     = testcase.Var{Name: `blk`}
			subject = func(t *testcase.T) bool {
				return customTBGet(t).Run(name.Get(t).(string), blk.Get(t).(func(testing.TB)))
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

				require.False(t, customTBGet(t).Failed())
			})
		})

		s.When(`block fails out early`, func(s *testcase.Spec) {
			blk.Let(s, func(t *testcase.T) interface{} {
				return func(tb testing.TB) { tb.FailNow() }
			})

			s.Then(`it will report the fail`, func(t *testcase.T) {
				require.False(t, subject(t))
			})

			s.Then(`it will mark the parent as failed`, func(t *testcase.T) {
				subject(t)

				require.True(t, customTBGet(t).Failed())
			})
		})
	})
}
