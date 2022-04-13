package contracts

import (
	"os"
	"sync"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/random"
)

type CustomTB struct {
	NewSubject func(testing.TB) testcase.TBRunner
}

func (spec CustomTB) Test(t *testing.T) { spec.Spec(t) }

func (spec CustomTB) Benchmark(b *testing.B) { spec.Spec(b) }

func (spec CustomTB) Spec(tb testing.TB) {
	s := testcase.NewSpec(tb)

	customTB := testcase.Let(s, func(t *testcase.T) interface{} {
		return spec.NewSubject(t)
	})
	customTBGet := func(t *testcase.T) testcase.TBRunner {
		return customTB.Get(t).(testcase.TBRunner)
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
		assert.Must(t).True(wasCancelled)
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

			assert.Must(t).True(customTBGet(t).Failed())
		})
	}

	thenItWillNotMarkTheTestAsFailed := func(s *testcase.Spec, subject func(t *testcase.T)) {
		s.Then(`it will not mark the test as failed`, func(t *testcase.T) {
			subject(t)

			assert.Must(t).True(!customTBGet(t).Failed())
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
			assert.Must(t).True(!subject(t))
		})

		s.When(`test failed`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				customTBGet(t).Fail()
			})

			s.Then(`it will be true`, func(t *testcase.T) {
				assert.Must(t).True(subject(t))
			})
		})
	})

	s.Describe(`#Log`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		var subject = func(t *testcase.T) {
			customTBGet(t).Log(rndInterfaceListArgs.Get(t)...)
		}

		thenItWillNotMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Logf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) {
			customTBGet(t).Logf(rndInterfaceListFormat.Get(t), rndInterfaceListArgs.Get(t)...)
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
			t.Must.NotEqual(0, len(subject(t)))
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
				customTBGet(t).Skip(rndInterfaceListArgs.Get(t)...)
			})
		}

		thenItWillNotMarkTheTestAsFailed(s, subject)
	})

	s.Describe(`#Skipf`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)
		var subject = func(t *testcase.T) {
			expectToExitGoroutine(t, func() {
				customTBGet(t).Skipf(rndInterfaceListFormat.Get(t), rndInterfaceListArgs.Get(t)...)
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
				assert.Must(t).True(subject(t))
			})
		})

		s.Test(`by default tests are not skipped so it will report false`, func(t *testcase.T) {
			assert.Must(t).True(!subject(t))
		})
	})

	s.Describe(`#TempDir`, func(s *testcase.Spec) {
		rndInterfaceListArgs.Let(s, nil)
		rndInterfaceListFormat.Let(s, nil)

		type TempDirer interface{ TempDir() string }
		var (
			getTempDirer = func(t *testcase.T) TempDirer {
				td, ok := customTB.Get(t).(TempDirer)
				if !ok {
					t.Skip(`testing.TB don't support TempDir() string method`)
				}
				return td
			}
			subject = func(t *testcase.T) string {
				return getTempDirer(t).TempDir()
			}
		)

		thenItWillNotMarkTheTestAsFailed(s, func(t *testcase.T) { subject(t) })

		s.Then(`return an existing directory`, func(t *testcase.T) {
			tmpdir := subject(t)

			t.Must.True(0 < len(tmpdir))
			if fi, err := os.Stat(tmpdir); err != nil {
				assert.Must(t).True(!os.IsNotExist(err), `expected to !os.IsNotExist`)
			} else {
				assert.Must(t).True(fi.Mode().IsDir())
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

			t.Must.Equal([]int{4, 2}, cleanups)
		})
	})

	s.Describe(`#Run`, func(s *testcase.Spec) {
		var (
			name    = testcase.LetValue(s, rnd.String())
			blk     = testcase.Var[func(testing.TB)]{ID: `blk`}
			subject = func(t *testcase.T) bool {
				return customTBGet(t).Run(name.Get(t), blk.Get(t))
			}
		)

		s.When(`block result in a passing sub test`, func(s *testcase.Spec) {
			blk.Let(s, func(t *testcase.T) func(testing.TB) {
				return func(testing.TB) {}
			})

			s.Then(`it will report the success`, func(t *testcase.T) {
				assert.Must(t).True(subject(t))
			})

			s.Then(`it will not mark the parent as failed`, func(t *testcase.T) {
				subject(t)

				assert.Must(t).True(!customTBGet(t).Failed())
			})
		})

		s.When(`block fails out early`, func(s *testcase.Spec) {
			blk.Let(s, func(t *testcase.T) func(testing.TB) {
				return func(tb testing.TB) { tb.FailNow() }
			})

			s.Then(`it will report the fail`, func(t *testcase.T) {
				assert.Must(t).True(!subject(t))
			})

			s.Then(`it will mark the parent as failed`, func(t *testcase.T) {
				subject(t)

				assert.Must(t).True(customTBGet(t).Failed())
			})
		})
	})
}
