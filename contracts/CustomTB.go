package contracts

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
)

type CustomTB struct {
	Subject func(*testcase.T) testcase.TBRunner
}

func (c CustomTB) Test(t *testing.T) { c.Spec(testcase.NewSpec(t)) }

func (c CustomTB) Benchmark(b *testing.B) { c.Spec(testcase.NewSpec(b)) }

func (c CustomTB) Spec(s *testcase.Spec) {
	customTB := testcase.Let(s, func(t *testcase.T) testcase.TBRunner {
		return c.Subject(t)
	})

	s.Describe(`#Run`, func(s *testcase.Spec) {
		var (
			name = testcase.Let(s, func(t *testcase.T) string {
				return t.Random.String()
			})
			blk     = testcase.Var[func(testing.TB)]{ID: `blk`}
			subject = func(t *testcase.T) bool {
				return customTB.Get(t).Run(name.Get(t), blk.Get(t))
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

				assert.Must(t).True(!customTB.Get(t).Failed())
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

				assert.Must(t).True(customTB.Get(t).Failed())
			})
		})
	})

	s.Context("implements testing.TB", func(s *testcase.Spec) {
		testcase.RunSuite(s, TestingTB{
			Subject: func(t *testcase.T) testing.TB {
				return c.Subject(t)
			},
		})
	})
}
