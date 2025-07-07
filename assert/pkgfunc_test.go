package assert_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/sandbox"
)

func TestPublicFunctions(t *testing.T) {
	type TestCase struct {
		Desc   string
		Failed bool
		Assert func(testing.TB)
	}

	for _, tc := range []TestCase{
		// .True
		{
			Desc:   ".True - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.True(tb, true)
			},
		},
		{
			Desc:   ".True - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.True(tb, false)
			},
		},
		// .False
		{
			Desc:   ".False - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.False(tb, false)
			},
		},
		{
			Desc:   ".False - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.False(tb, true)
			},
		},
		// .Nil
		{
			Desc:   ".Nil - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.Nil(tb, nil)
			},
		},
		{
			Desc:   ".Nil - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.Nil(tb, &TestCase{})
			},
		},
		// .NotNil
		{
			Desc:   ".NotNil - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.NotNil(tb, &TestCase{})
			},
		},
		{
			Desc:   ".NotNil - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.NotNil(tb, nil)
			},
		},
		// .Empty
		{
			Desc:   ".Empty - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.Empty(tb, []int{})
			},
		},
		{
			Desc:   ".Empty - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.Empty(tb, []int{42})
			},
		},
		// .NotEmpty
		{
			Desc:   ".NotEmpty - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.NotEmpty(tb, []int{42})
			},
		},
		{
			Desc:   ".NotEmpty - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.NotEmpty(tb, []int{})
			},
		},
		// .Panic
		{
			Desc:   ".Panic - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				expected := "boom!"
				actual := assert.Panic(tb, func() { panic(expected) })
				assert.Must(tb).Equal(expected, actual)
			},
		},
		{
			Desc:   ".Panic - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.Panic(tb, func() {})
			},
		},
		// .NotPanic
		{
			Desc:   ".NotPanic - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.NotPanic(tb, func() {})
			},
		},
		{
			Desc:   ".NotPanic - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.NotPanic(tb, func() { panic("boom!") })
			},
		},
		// .Equal
		{
			Desc:   ".Equal - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.Equal(tb, 42, 42)
				assert.Equal(tb, "42", "42")
				assert.Equal(tb, []string{"42"}, []string{"42"})
			},
		},
		{
			Desc:   ".Equal - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.Equal(tb, 42, 24)
			},
		},
		// .NotEqual
		{
			Desc:   ".NotEqual - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.NotEqual(tb, 42, 24)
				assert.NotEqual(tb, "42", "24")
				assert.NotEqual(tb, []string{"42"}, []string{"42", "24"})
			},
		},
		{
			Desc:   ".NotEqual - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.NotEqual(tb, 42, 42)
			},
		},
		// .Contain
		{
			Desc:   ".Contain - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.Contain(tb, "The Answer is 42", "42")
				assert.Contain(tb, []string{"42", "24"}, "42")
				assert.Contain(tb, map[string]int{"The answer": 42, "Are you good?": 0}, map[string]int{"The answer": 42})
			},
		},
		{
			Desc:   ".Contain - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.Contain(tb, "The Answer is 42", "422")
			},
		},
		// .NotContain
		{
			Desc:   ".NotContain - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.NotContain(tb, "The Answer is 42", "422")
				assert.NotContain(tb, []string{"42", "24"}, "13")
				assert.NotContain(tb,
					map[string]int{"The answer": 42, "Are you good?": 0},
					map[string]int{"The answer to you": 42})
			},
		},
		{
			Desc:   ".NotContain - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.NotContain(tb, "The Answer is 42", "42")
			},
		},
		// .ContainExactly
		{
			Desc:   ".ContainExactly - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.ContainExactly(tb, []string{"42", "24"}, []string{"24", "42"})
				assert.ContainExactly(tb, map[string]int{"a": 1, "b": 2}, map[string]int{"b": 2, "a": 1})
			},
		},
		{
			Desc:   ".ContainExactly - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.ContainExactly(tb, []string{"42", "24"}, []string{"24", "42", "13"})
			},
		},
		// .Subset
		{
			Desc:   ".Subset - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.Sub(tb, []string{"42", "24", "13"}, []string{"42", "24"})
			},
		},
		{
			Desc:   ".Subset - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.Sub(tb, []string{"42", "24", "13"}, []string{"24", "42"})
			},
		},
		// .ErrorIs
		{
			Desc:   ".ErrorIs - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				expected := errors.New("boom")
				actual := fmt.Errorf("wrapped boom: %w", expected)
				assert.ErrorIs(tb, expected, actual)
			},
		},
		{
			Desc:   ".ErrorIs - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				expected := errors.New("boom")
				actual := fmt.Errorf("wrapped boom: %v", expected)
				assert.ErrorIs(tb, expected, actual)
			},
		},
		// .Error
		{
			Desc:   ".Error - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.Error(tb, errors.New("boom"))
			},
		},
		{
			Desc:   ".NoError - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.Error(tb, nil)
			},
		},
		// .NoError
		{
			Desc:   ".NoError - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.NoError(tb, nil)
			},
		},
		{
			Desc:   ".NoError - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.NoError(tb, errors.New("boom"))
			},
		},
		// Read
		{
			Desc:   ".Read - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.Read(tb, "foo", strings.NewReader("foo"))
			},
		},
		{
			Desc:   ".Read - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.Read(tb, "bar", strings.NewReader("foo"))
			},
		},
		// ReadAll
		{
			Desc:   ".ReadAll - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				expected := "foo"
				assert.Equal(tb, expected, string(assert.ReadAll(tb, strings.NewReader(expected))))
			},
		},
		{
			Desc:   ".ReadAll - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.ReadAll(tb, iotest.ErrReader(errors.New("boom")))
			},
		},
		// Within
		{
			Desc:   ".Within - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.Within(tb, time.Second, func(ctx context.Context) {})
			},
		},
		{
			Desc:   ".Within - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.Within(tb, time.Nanosecond, func(ctx context.Context) { time.Sleep(time.Second) })
			},
		},
		// Not Within
		{
			Desc:   ".NotWithin - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.NotWithin(tb, time.Nanosecond, func(ctx context.Context) {
					time.Sleep(time.Millisecond)
				})
			},
		},
		{
			Desc:   ".NotWithin - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.NotWithin(tb, time.Millisecond, func(ctx context.Context) {
					// time.Sleep(time.Nanosecond)
				})
			},
		},
		// .Match
		{
			Desc:   ".Match - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.MatchRegexp(tb, "42", "[0-9]+")
				assert.MatchRegexp(tb, "forty-two", "[a-z]+")
				assert.MatchRegexp(tb, []byte("forty-two"), "[a-z]+")
			},
		},
		{
			Desc:   ".Match - happy - subtype",
			Failed: false,
			Assert: func(tb testing.TB) {
				type S string
				assert.MatchRegexp(tb, S("42"), "[0-9]+")
				assert.MatchRegexp(tb, S("forty-two"), "[a-z]+")
			},
		},
		{
			Desc:   ".Match - rainy value",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.MatchRegexp(tb, "42", "[a-z]+")
			},
		},
		{
			Desc:   ".Match - rainy pattern",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.MatchRegexp(tb, "42", "[0-9")
			},
		},
		// .NotMatch
		{
			Desc:   ".NotMatch - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.NotMatchRegexp(tb, "forty-two", "^[0-9]+")
				assert.NotMatchRegexp(tb, "42", "^[a-z]+")
				assert.NotMatchRegexp(tb, []byte("forty-two"), "^[0-9]+")
			},
		},
		{
			Desc:   ".NotMatch - happy - subtype",
			Failed: false,
			Assert: func(tb testing.TB) {
				type S string
				assert.NotMatchRegexp(tb, S("forty-two"), "^[0-9]+")
				assert.NotMatchRegexp(tb, S("42"), "^[a-z]+")
			},
		},
		{
			Desc:   ".NotMatch - rainy value",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.NotMatchRegexp(tb, "42", "[0-9]+")
			},
		},
		{
			Desc:   ".NotMatch - rainy pattern",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.NotMatchRegexp(tb, "forty-two", "[0-9")
			},
		},
		// .Eventually
		{
			Desc:   ".Eventually - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				var ok bool
				assert.Eventually(tb, 2, func(it testing.TB) {
					if ok {
						return
					}
					ok = true
					it.FailNow()
				})
			},
		},
		{
			Desc:   ".Eventually - rainy value",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.Eventually(tb, 1, func(it testing.TB) {
					it.FailNow()
				})
			},
		},
		// .AnyOf
		{
			Desc:   ".AnyOf - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.AnyOf(tb, func(a *assert.A) {
					a.Case(func(it testing.TB) { it.FailNow() })
					a.Case(func(it testing.TB) {})
				})
			},
		},
		{
			Desc:   ".AnyOf - rainy value",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.AnyOf(tb, func(a *assert.A) {
					a.Case(func(it testing.TB) { it.FailNow() })
					a.Case(func(it testing.TB) { it.FailNow() })
				})
			},
		},
		// .Unique
		{
			Desc:   ".Unique - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.Unique(tb, []int{1, 2, 3})
			},
		},
		{
			Desc:   ".Unique - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.Unique(tb, []int{1, 2, 3, 4, 1})
			},
		},
		// .NotUnique
		{
			Desc:   ".NotUnique - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.NotUnique(tb, []int{1, 2, 3, 1})
			},
		},
		{
			Desc:   ".NotUnique - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.NotUnique(tb, []int{1, 2, 3, 4})
			},
		},
		// .OneOf
		{
			Desc:   ".OneOf - happy",
			Failed: false,
			Assert: func(tb testing.TB) {
				assert.OneOf[int](tb, []int{1, 2, 3}, func(t testing.TB, got int) {
					assert.Equal(t, 3, got)
				})
			},
		},
		{
			Desc:   ".OneOf - rainy",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.OneOf[int](tb, []int{1, 2, 3}, func(t testing.TB, got int) {
					assert.Equal(t, 4, got)
				})
			},
		},
	} {
		t.Run(tc.Desc, func(t *testing.T) {
			stub := &doubles.TB{}
			out := sandbox.Run(func() {
				tc.Assert(stub)
			})
			assert.Must(t).Equal(tc.Failed, stub.IsFailed, "expected / got")
			if tc.Failed {
				assert.Must(t).False(out.OK, "Test was expected to fail with Fatal/FailNow")
			} else {
				assert.True(t, 0 < stub.Passes())
			}
		})
	}
}
