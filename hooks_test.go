package testcase_test

import (
	"strconv"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/internal"
	"github.com/stretchr/testify/require"
)

func TestSpec_Before_Ordered(t *testing.T) {
	var (
		actually []int
		expected []int
	)

	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.Sequential()

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

	require.Equal(t, expected, actually)
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

	require.Equal(t, []int{6, 5, 4, 3, 2, 1}, afters)
}

func TestSpec_BeforeAll_blkRunsOnlyOnce(t *testing.T) {
	s := testcase.NewSpec(t)

	var counter int
	blk := func(t *testcase.T) { require.Equal(t, 1, counter) }
	s.BeforeAll(func(tb testing.TB) { counter++ })
	s.Test(``, blk)
	s.Test(``, blk)
	s.Test(``, blk)
	s.Context(``, func(s *testcase.Spec) {
		s.Test(``, blk)
		s.Test(``, blk)
		s.Test(``, blk)
	})
	s.Finish()

	require.Equal(t, 1, counter)
}

func TestSpec_AfterAll_blkRunsOnlyOnce(t *testing.T) {
	s := testcase.NewSpec(t)

	var counter int
	blk := func(t *testcase.T) { require.Equal(t, 0, counter) }
	s.AfterAll(func(tb testing.TB) { counter++ })
	s.Test(``, blk)
	s.Test(``, blk)
	s.Test(``, blk)
	s.Context(``, func(s *testcase.Spec) {
		s.Test(``, blk)
		s.Test(``, blk)
		s.Test(``, blk)
	})
	s.Finish()

	require.Equal(t, 1, counter)
}

func TestSpec_AroundAll_blkRunsOnlyOnce(t *testing.T) {
	s := testcase.NewSpec(t)

	var before, after int
	blk := func(t *testcase.T) {
		require.Equal(t, 1, before)
		require.Equal(t, 0, after)
	}
	s.AroundAll(func(tb testing.TB) func() {
		before++
		return func() { after++ }
	})
	s.Test(``, blk)
	s.Test(``, blk)
	s.Test(``, blk)
	s.Context(``, func(s *testcase.Spec) {
		s.Test(``, blk)
		s.Test(``, blk)
		s.Test(``, blk)
	})
	s.Finish()

	require.Equal(t, 1, before)
	require.Equal(t, 1, after)
}

func TestSpec_BeforeAll_failIfDefinedAfterTestCases(t *testing.T) {
	var isAnyOfTheTestCaseRan bool
	blk := func(t *testcase.T) { isAnyOfTheTestCaseRan = true }
	stub := &internal.StubTB{}

	internal.InGoroutine(func() {
		s := testcase.NewSpec(stub)
		s.Test(``, blk)
		s.BeforeAll(func(tb testing.TB) {})
		s.Test(``, blk)
		s.Finish()
	})

	require.True(t, stub.IsFailed)
	require.False(t, isAnyOfTheTestCaseRan)
}

func TestSpec_AfterAll_failIfDefinedAfterTestCases(t *testing.T) {
	var isAnyOfTheTestCaseRan bool
	blk := func(t *testcase.T) { isAnyOfTheTestCaseRan = true }
	stub := &internal.StubTB{}

	internal.InGoroutine(func() {
		s := testcase.NewSpec(stub)
		s.Test(``, blk)
		s.AfterAll(func(tb testing.TB) {})
		s.Test(``, blk)
		s.Finish()
	})

	require.True(t, stub.IsFailed)
	require.False(t, isAnyOfTheTestCaseRan)
}

func TestSpec_AroundAll_failIfDefinedAfterTestCases(t *testing.T) {
	var isAnyOfTheTestCaseRan bool
	blk := func(t *testcase.T) { isAnyOfTheTestCaseRan = true }
	stub := &internal.StubTB{}

	internal.InGoroutine(func() {
		s := testcase.NewSpec(stub)
		s.Test(``, blk)
		s.AroundAll(func(tb testing.TB) func() { return func() {} })
		s.Test(``, blk)
		s.Finish()
	})

	require.True(t, stub.IsFailed)
	require.False(t, isAnyOfTheTestCaseRan)
}
