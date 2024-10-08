package testcase_test

import (
	"strconv"
	"testing"

	"go.llib.dev/testcase/sandbox"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
)

func TestSpec_Before_Ordered(t *testing.T) {
	var (
		actually []int
		expected []int
	)

	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.Sequential()

		s.Context("", func(s *testcase.Spec) {
			last := s
			for i := 0; i < 5; i++ {
				currentValue := i
				expected = append(expected, currentValue)

				last.Context(strconv.Itoa(currentValue), func(next *testcase.Spec) {
					next.Before(func(t *testcase.T) {
						actually = append(actually, currentValue)
					})
					last = next
				})
			}

			last.Test(`trigger hooks now`, func(t *testcase.T) {})
		})
	})

	assert.Must(t).Equal(expected, actually)
}

func TestSpec_After(t *testing.T) {
	var afters []int

	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)

		s.After(func(t *testcase.T) { afters = append(afters, 1) })
		s.After(func(t *testcase.T) { afters = append(afters, 2) })
		s.After(func(t *testcase.T) { afters = append(afters, 3) })

		s.Context(`in spec`, func(s *testcase.Spec) {
			s.After(func(t *testcase.T) { afters = append(afters, 4) })
			s.After(func(t *testcase.T) { afters = append(afters, 5) })
			s.After(func(t *testcase.T) { afters = append(afters, 6) })
			s.Test(`in testCase`, func(t *testcase.T) {})
		})
	})

	assert.Must(t).Equal([]int{6, 5, 4, 3, 2, 1}, afters)
}

func TestSpec_BeforeAll_blkRunsOnlyOnce(t *testing.T) {
	s := testcase.NewSpec(t)

	var counter int
	blk := func(t *testcase.T) { assert.Must(t).Equal(1, counter) }
	s.BeforeAll(func(tb testing.TB) { counter++ })
	s.Test(``, blk)
	s.Test(``, blk)
	s.Test(``, blk)
	s.Context(``, func(s *testcase.Spec) {
		s.Test(``, blk)
		s.Test(``, blk)
		s.Test(``, blk)
	})

	assert.Must(t).Equal(1, counter)
}

func TestSpec_BeforeAll_failIfDefinedAfterTestCases(t *testing.T) {
	stub := &doubles.TB{}
	sandbox.Run(func() {
		s := testcase.NewSpec(stub)
		s.Test(``, func(t *testcase.T) {})
		s.BeforeAll(func(tb testing.TB) {})
		s.Test(``, func(t *testcase.T) {})
		s.Finish()
	})
	assert.Must(t).True(stub.IsFailed)
}

func ExampleSpec_AfterAll() {
	s := testcase.NewSpec(nil)
	s.AfterAll(func(tb testing.TB) {
		// do something after all the test finished running
	})
	s.Test("this test will run before the AfterAll hook", func(t *testcase.T) {})
}

func TestSpec_AfterAll(t *testing.T) {
	stub := &doubles.TB{}
	var order []string
	sandbox.Run(func() {
		s := testcase.NewSpec(stub)
		s.HasSideEffect()
		s.AfterAll(func(tb testing.TB) {
			order = append(order, "AfterAll")
		})
		s.Test(``, func(t *testcase.T) {
			order = append(order, "Test")
		})
		s.Test(``, func(t *testcase.T) {
			order = append(order, "Test")
		})
		s.Finish()
	})
	assert.Must(t).False(stub.IsFailed)
	assert.Equal(t, []string{"Test", "Test", "AfterAll"}, order,
		`expected to only run once (single "AfterAll" in the order array)`,
		`and it should have run in order (After all the "Test")`,
	)
}

func TestSpec_AfterAll_nested(t *testing.T) {
	stub := &doubles.TB{}
	var order []string
	sandbox.Run(func() {
		s := testcase.NewSpec(stub)
		s.HasSideEffect()
		s.AfterAll(func(tb testing.TB) {
			order = append(order, "AfterAll")
		})
		s.Context(``, func(s *testcase.Spec) {
			s.AfterAll(func(tb testing.TB) {
				order = append(order, "AfterAll")
			})
			s.Test(``, func(t *testcase.T) {
				order = append(order, "Test")
			})
		})
		s.Test(``, func(t *testcase.T) {
			order = append(order, "Test")
		})
		s.Finish()
	})
	assert.Must(t).False(stub.IsFailed)
	assert.Equal(t, []string{"Test", "Test", "AfterAll", "AfterAll"}, order)
}

func TestSpec_AfterAll_suite(t *testing.T) {
	stub := &doubles.TB{}
	var order []string
	sandbox.Run(func() {
		suiteSpec1 := testcase.NewSpec(nil)
		suiteSpec1.HasSideEffect()
		suiteSpec1.AfterAll(func(tb testing.TB) {
			order = append(order, "AfterAll")
		})
		suiteSpec1.Test(``, func(t *testcase.T) {
			order = append(order, "Test")
		})
		suite1 := suiteSpec1.AsSuite("suite")
		suiteSpec2 := testcase.NewSpec(nil)
		suiteSpec2.Context("", suite1.Spec)

		ss := testcase.NewSpec(stub)
		ss.Context("", suiteSpec2.Spec)
		ss.Finish()
	})
	assert.Must(t).False(stub.IsFailed)
	assert.Equal(t, []string{"Test", "AfterAll"}, order, "expected to only run once, in the real spec execution")
}

func TestSpec_AfterAll_failIfDefinedAfterTestCases(t *testing.T) {
	stub := &doubles.TB{}
	sandbox.Run(func() {
		s := testcase.NewSpec(stub)
		s.Test(``, func(t *testcase.T) {})
		s.AfterAll(func(tb testing.TB) {})
		s.Test(``, func(t *testcase.T) {})
		s.Finish()
	})
	assert.Must(t).True(stub.IsFailed)
}
