package assert_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/sandbox"
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
		// .ErrorIs
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
			Desc:   ".Read - rany",
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
			Desc:   ".ReadAll - rany",
			Failed: true,
			Assert: func(tb testing.TB) {
				assert.ReadAll(tb, iotest.ErrReader(errors.New("boom")))
			},
		},
	} {
		t.Run(tc.Desc, func(t *testing.T) {
			stub := &testcase.StubTB{}
			sandbox.Run(func() {
				tc.Assert(stub)
			})
			assert.Must(t).Equal(tc.Failed, stub.IsFailed)
		})
	}
}
