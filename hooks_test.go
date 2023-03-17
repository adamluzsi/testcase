package testcase_test

import (
	"strconv"
	"testing"

	"github.com/adamluzsi/testcase/sandbox"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/doubles"
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
