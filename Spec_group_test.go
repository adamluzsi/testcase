package testcase_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
)

// rendezvous returns a function that blocks until n participants called it.
// If the participants can't meet, because they don't run concurrently, it fails the test.
func rendezvous(n int) func(testing.TB) {
	var wg sync.WaitGroup
	wg.Add(n)
	return func(tb testing.TB) {
		tb.Helper()
		wg.Done()
		met := make(chan struct{})
		go func() {
			defer close(met)
			wg.Wait()
		}()
		select {
		case <-met:
		case <-time.After(5 * time.Second):
			tb.Fatal("tests were expected to run concurrently, but they ran sequentially")
		}
	}
}

func TestSpec_Context_groupParallelism(t *testing.T) {
	const n = 3
	suite := func(name string, noSideEffect bool, test func(t *testcase.T)) testcase.SpecSuite {
		s := testcase.NewSpec(nil)
		if noSideEffect {
			s.NoSideEffect()
		}
		s.Test("", test)
		return s.AsSuite(name)
	}

	t.Run("when the suites declare that they have no side effect, they run in parallel", func(t *testing.T) {
		meet := rendezvous(n)
		s := testcase.NewSpec(t)
		for i := 0; i < n; i++ {
			name := strconv.Itoa(i)
			s.Context(name, suite(name, true, func(t *testcase.T) { meet(t) }).Spec)
		}
	})

	t.Run("when the outer spec declares that it has no side effect, the suites run in parallel", func(t *testing.T) {
		meet := rendezvous(n)
		s := testcase.NewSpec(t)
		s.NoSideEffect()
		for i := 0; i < n; i++ {
			name := strconv.Itoa(i)
			s.Context(name, suite(name, false, func(t *testcase.T) { meet(t) }).Spec)
		}
	})

	t.Run("when the suites are executed with RunSuite, they run in parallel", func(t *testing.T) {
		meet := rendezvous(n)
		var suites []testcase.Suite
		for i := 0; i < n; i++ {
			suites = append(suites, suite(strconv.Itoa(i), true, func(t *testcase.T) { meet(t) }))
		}
		testcase.RunSuite(t, suites...)
	})

	t.Run("when a parallel group is nested in another parallel group, they run in parallel", func(t *testing.T) {
		meet := rendezvous(n)
		s := testcase.NewSpec(t)
		s.NoSideEffect()
		s.Context("outer", func(s *testcase.Spec) {
			s.Test("", func(t *testcase.T) { meet(t) })
			s.Context("inner", func(s *testcase.Spec) {
				s.Test("", func(t *testcase.T) { meet(t) })
			}, testcase.Group("inner"))
		}, testcase.Group("outer"))
		s.Test("", func(t *testcase.T) { meet(t) })
	})

	t.Run("when a group has side effect somewhere in it, the group runs sequentially", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.NoSideEffect()
		s.Context("A", func(s *testcase.Spec) {
			s.Test("", func(t *testcase.T) {})
			s.Context("with side effect", func(s *testcase.Spec) {
				s.HasSideEffect()
				s.Test("", func(t *testcase.T) {
					// testing.T#Setenv panics when the test or any of its parents is parallel.
					assert.NotPanic(t, func() { t.Setenv("TESTCASE_GROUP_PARALLELISM", "A") })
				})
			})
		}, testcase.Group("A"))
		s.Context("B", func(s *testcase.Spec) {
			s.Test("", func(t *testcase.T) {})
		}, testcase.Group("B"))
	})

	t.Run("AfterAll runs after the tests of the parallel groups", func(t *testing.T) {
		type Recorder struct {
			m     sync.Mutex
			order []string
		}
		record := func(r *Recorder, v string) {
			r.m.Lock()
			defer r.m.Unlock()
			r.order = append(r.order, v)
		}
		parallelSuite := func(r *Recorder) testcase.SpecSuite {
			s := testcase.NewSpec(nil)
			s.NoSideEffect()
			s.AfterAll(func(tb testing.TB) { record(r, "AfterAll") })
			s.Test("", func(t *testcase.T) { record(r, "Test") })
			s.Test("", func(t *testcase.T) { record(r, "Test") })
			return s.AsSuite("suite")
		}

		t.Run("with SpecSuite#Test", func(t *testing.T) {
			var r Recorder
			t.Run("", func(t *testing.T) { parallelSuite(&r).Test(t) })
			assert.Equal(t, []string{"Test", "Test", "AfterAll"}, r.order)
		})

		t.Run("with RunSuite", func(t *testing.T) {
			var r Recorder
			t.Run("", func(t *testing.T) { testcase.RunSuite(t, parallelSuite(&r)) })
			assert.Equal(t, []string{"Test", "Test", "AfterAll"}, r.order)
		})

		t.Run("with an explicit Spec#Finish", func(t *testing.T) {
			var r Recorder
			t.Run("", func(t *testing.T) {
				s := testcase.NewSpec(t)
				s.NoSideEffect()
				s.AfterAll(func(tb testing.TB) { record(&r, "AfterAll") })
				s.Test("", func(t *testcase.T) { record(&r, "Test") })
				s.Test("", func(t *testcase.T) { record(&r, "Test") })
				s.Finish()
			})
			assert.Equal(t, []string{"Test", "Test", "AfterAll"}, r.order)
		})
	})
}
