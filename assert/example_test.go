package assert_test

import (
	"errors"
	"testing"

	"github.com/adamluzsi/testcase/assert"
)

func ExampleMust() {
	var tb testing.TB
	// create an assertion helper which will fail the testing context with .Fatal(...) in case of a failed assert.
	assert.Must(tb).True(true)
}

func ExampleShould() {
	var tb testing.TB
	// create an assertion helper which will fail the testing context with .Error(...) in case of a failed assert.
	assert.Should(tb).True(true)
}

func ExampleAsserter_True() {
	var tb testing.TB
	assert.Must(tb).True(true, "optional assertion explanation")
}

func ExampleAsserter_Nil() {
	var tb testing.TB
	assert.Must(tb).Nil(nil, "optional assertion explanation")
}

func ExampleAsserter_NotNil() {
	var tb testing.TB
	assert.Must(tb).NotNil(errors.New("42"), "optional assertion explanation")
}

func ExampleAsserter_Equal() {
	var tb testing.TB
	assert.Must(tb).Equal(true, true, "optional assertion explanation")
}

func ExampleAsserter_NotEqual() {
	var tb testing.TB
	assert.Must(tb).NotEqual(true, false, "optional assertion explanation")
}

func ExampleAsserter_Contain() {
	var tb testing.TB
	assert.Must(tb).Contain([]int{1, 2, 3}, 3, "optional assertion explanation")
	assert.Must(tb).Contain([]int{1, 2, 3}, []int{1, 2}, "optional assertion explanation")
	assert.Must(tb).Contain(map[string]int{"The Answer": 42, "oth": 13}, map[string]int{"The Answer": 42}, "optional assertion explanation")
}

func ExampleAsserter_NotContain() {
	var tb testing.TB
	assert.Must(tb).NotContain([]int{1, 2, 3}, 42, "optional assertion explanation")
	assert.Must(tb).NotContain([]int{1, 2, 3}, []int{42}, "optional assertion explanation")
	assert.Must(tb).NotContain(map[string]int{"The Answer": 42, "oth": 13}, map[string]int{"The Answer": 13}, "optional assertion explanation")
}

func ExampleAsserter_ContainExactly() {
	var tb testing.TB
	assert.Must(tb).ContainExactly([]int{1, 2, 3}, []int{2, 3, 1}, "optional assertion explanation")  // true
	assert.Must(tb).ContainExactly([]int{1, 2, 3}, []int{1, 42, 2}, "optional assertion explanation") // false
}

func ExampleAsserter_Panic() {
	var tb testing.TB
	assert.Must(tb).Panic(func() { panic("boom") }, "optional assertion explanation")
}

func ExampleAsserter_NotPanic() {
	var tb testing.TB
	assert.Must(tb).NotPanic(func() { /* no boom */ }, "optional assertion explanation")
}

func ExampleAsserter_AnyOf() {
	var tb testing.TB
	var list []interface {
		Foo() int
		Bar() bool
		Baz() string
	}
	assert.Must(tb).AnyOf(func(anyOf *assert.AnyOf) {
		for _, testingCase := range list {
			anyOf.Test(func(it assert.It) {
				it.Must.True(testingCase.Bar())
			})
		}
	})
}

func ExampleAnyOf() {
	var tb testing.TB
	var list []interface {
		Foo() int
		Bar() bool
		Baz() string
	}
	anyOf := assert.AnyOf{TB: tb, Fn: tb.Fatal}
	for _, testingCase := range list {
		anyOf.Test(func(it assert.It) {
			it.Must.True(testingCase.Bar())
		})
	}
	anyOf.Finish()
}
