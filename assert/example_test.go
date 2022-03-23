package assert_test

import (
	"errors"
	"fmt"
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

func ExampleAsserter_False() {
	var tb testing.TB
	assert.Must(tb).False(false, "optional assertion explanation")
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

func ExampleAnyOf_listOfInterface() {
	var tb testing.TB
	type ExampleInterface interface {
		Foo() int
		Bar() bool
		Baz() string
	}
	anyOf := assert.AnyOf{TB: tb, Fn: tb.Fatal}
	for _, v := range []ExampleInterface{} {
		anyOf.Test(func(it assert.It) {
			it.Must.True(v.Bar())
		})
	}
	anyOf.Finish()
}

func ExampleAnyOf_listOfCompositedStructuresWhereOnlyTheEmbededValueIsRelevant() {
	var tb testing.TB
	type BigStruct struct {
		ID            string // not relevant for the test
		A, B, C, D, E int    // not relevant data as well
		WrappedStruct struct {
			A, B, C int // relevant data for the test
		}
	}
	anyOf := assert.AnyOf{TB: tb, Fn: tb.Fatal}
	for _, v := range []BigStruct{} {
		anyOf.Test(func(it assert.It) {
			it.Must.Equal(42, v.WrappedStruct.A)
			it.Must.Equal(1, v.WrappedStruct.B)
			it.Must.Equal(2, v.WrappedStruct.C)
		})
	}
	anyOf.Finish()
}

func ExampleAnyOf_listOfStructuresWithIrrelevantValues() {
	var tb testing.TB
	type StructWithDynamicValues struct {
		IrrelevantStateValue int // not relevant data for the test
		ImportantValue       int
	}
	anyOf := assert.AnyOf{TB: tb, Fn: tb.Fatal}
	for _, v := range []StructWithDynamicValues{} {
		anyOf.Test(func(it assert.It) {
			it.Must.Equal(42, v.ImportantValue)
		})
	}
	anyOf.Finish()
}

func ExampleAnyOf_structWithManyAcceptableState() {
	var tb testing.TB
	type ExampleStruct struct {
		Type    string
		A, B, C int
	}
	var es ExampleStruct
	anyOf := assert.AnyOf{TB: tb, Fn: tb.Fatal}
	anyOf.Test(func(it assert.It) {
		it.Must.Equal(`foo`, es.Type)
		it.Must.Equal(1, es.A)
		it.Must.Equal(2, es.B)
		it.Must.Equal(3, es.C)
	})
	anyOf.Test(func(it assert.It) {
		it.Must.Equal(`foo`, es.Type)
		it.Must.Equal(3, es.A)
		it.Must.Equal(2, es.B)
		it.Must.Equal(1, es.C)
	})
	anyOf.Test(func(it assert.It) {
		it.Must.Equal(`bar`, es.Type)
		it.Must.Equal(11, es.A)
		it.Must.Equal(12, es.B)
		it.Must.Equal(13, es.C)
	})
	anyOf.Test(func(it assert.It) {
		it.Must.Equal(`baz`, es.Type)
		it.Must.Equal(21, es.A)
		it.Must.Equal(22, es.B)
		it.Must.Equal(23, es.C)
	})
	anyOf.Finish()
}

type ExamplePublisherEvent struct{ V int }
type ExamplePublisher struct{}

func (ExamplePublisher) Publish(event ExamplePublisherEvent)         {}
func (ExamplePublisher) Subscribe(func(event ExamplePublisherEvent)) {}
func (ExamplePublisher) Wait()                                       {}
func (ExamplePublisher) Close() error                                { return nil }

func ExampleAnyOf_fanOutPublishing() {
	var tb testing.TB
	publisher := ExamplePublisher{}
	anyOf := &assert.AnyOf{TB: tb, Fn: tb.Fatal}
	for i := 0; i < 42; i++ {
		publisher.Subscribe(func(event ExamplePublisherEvent) {
			anyOf.Test(func(it assert.It) {
				it.Must.Equal(42, event.V)
			})
		})
	}
	publisher.Publish(ExamplePublisherEvent{V: 42})
	publisher.Wait()
	assert.Must(tb).Nil(publisher.Close())
	anyOf.Finish()
}

func ExampleAsserter_Empty() {
	var tb testing.TB

	assert.Must(tb).Empty([]int{})   // pass
	assert.Must(tb).Empty([]int{42}) // fail

	assert.Must(tb).Empty([42]int{})   // pass
	assert.Must(tb).Empty([42]int{42}) // fail

	assert.Must(tb).Empty(map[int]int{})       // pass
	assert.Must(tb).Empty(map[int]int{42: 24}) // fail

	assert.Must(tb).Empty("")   // pass
	assert.Must(tb).Empty("42") // fail
}

func ExampleAsserter_NotEmpty() {
	var tb testing.TB

	assert.Must(tb).NotEmpty([]int{42}, "optional assertion explanation")

	assert.Must(tb).NotEmpty([]int{})   // fail
	assert.Must(tb).NotEmpty([]int{42}) // pass

	assert.Must(tb).NotEmpty([42]int{})   // fail
	assert.Must(tb).NotEmpty([42]int{42}) // pass

	assert.Must(tb).NotEmpty(map[int]int{})       // fail
	assert.Must(tb).NotEmpty(map[int]int{42: 24}) // pass

	assert.Must(tb).NotEmpty("")   // fail
	assert.Must(tb).NotEmpty("42") // pass
}

func ExampleAsserter_ErrorIs() {
	var tb testing.TB

	actualErr := errors.New("boom")
	assert.Must(tb).ErrorIs(errors.New("boom"), actualErr)                                  // passes for equality
	assert.Must(tb).ErrorIs(errors.New("boom"), fmt.Errorf("wrapped error: %w", actualErr)) // passes for wrapped errors
}
