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

type ExampleEqualable struct {
	IrrelevantExportedField int
	relevantUnexportedValue int
}

func (es ExampleEqualable) IsEqual(oth ExampleEqualable) bool {
	return es.relevantUnexportedValue == oth.relevantUnexportedValue
}

func ExampleAsserter_Equal_isEqualFunctionUsedForComparison() {
	var tb testing.TB

	expected := ExampleEqualable{
		IrrelevantExportedField: 42,
		relevantUnexportedValue: 24,
	}

	actual := ExampleEqualable{
		IrrelevantExportedField: 4242,
		relevantUnexportedValue: 24,
	}

	assert.Must(tb).Equal(expected, actual) // passes as by IsEqual terms the two value is equal
}

type ExampleEqualableWithError struct {
	IrrelevantExportedField int
	relevantUnexportedValue int
	IsEqualErr              error
}

func (es ExampleEqualableWithError) IsEqual(oth ExampleEqualableWithError) (bool, error) {
	return es.relevantUnexportedValue == oth.relevantUnexportedValue, es.IsEqualErr
}

func ExampleAsserter_Equal_isEqualFunctionThatSupportsErrorReturning() {
	var tb testing.TB

	expected := ExampleEqualableWithError{
		IrrelevantExportedField: 42,
		relevantUnexportedValue: 24,
		IsEqualErr:              errors.New("sadly something went wrong"),
	}

	actual := ExampleEqualableWithError{
		IrrelevantExportedField: 42,
		relevantUnexportedValue: 24,
	}

	assert.Must(tb).Equal(expected, actual) // fails because the error returned from the IsEqual function.
}

func ExampleTrue() {
	var tb testing.TB
	assert.True(tb, true)  // ok
	assert.True(tb, false) // Fatal
}

func ExampleFalse() {
	var tb testing.TB
	assert.False(tb, false) // ok
	assert.False(tb, true)  // Fatal
}

func ExampleNil() {
	var tb testing.TB
	assert.Nil(tb, nil)                // ok
	assert.Nil(tb, errors.New("boom")) // Fatal
}

func ExampleNotNil() {
	var tb testing.TB
	assert.NotNil(tb, errors.New("boom")) // ok
	assert.NotNil(tb, nil)                // Fatal
}

func ExampleEmpty() {
	var tb testing.TB
	assert.Empty(tb, "")       // ok
	assert.Empty(tb, "oh no!") // Fatal
}

func ExampleNotEmpty() {
	var tb testing.TB
	assert.NotEmpty(tb, "huh...") // ok
	assert.NotEmpty(tb, "")       // Fatal
}

func ExamplePanic() {
	var tb testing.TB

	panicValue := assert.Panic(tb, func() { panic("at the disco") }) // ok
	assert.Equal(tb, "some expected panic value", panicValue)

	assert.Panic(tb, func() {}) // Fatal
}

func ExampleNotPanic() {
	var tb testing.TB
	assert.NotPanic(tb, func() {})                  // ok
	assert.NotPanic(tb, func() { panic("oh no!") }) // Fatal
}

func ExampleEqual() {
	var tb testing.TB
	assert.Equal(tb, "a", "a")
	assert.Equal(tb, 42, 42)
	assert.Equal(tb, []int{42}, []int{42})
	assert.Equal(tb, map[int]int{24: 42}, map[int]int{24: 42})
}

func ExampleNotEqual() {
	var tb testing.TB
	assert.NotEqual(tb, "a", "b")
	assert.Equal(tb, 13, 42)
}

func ExampleContain() {
	var tb testing.TB
	assert.Must(tb).Contain(tb, []int{1, 2, 3}, 3, "optional assertion explanation")
	assert.Must(tb).Contain(tb, []int{1, 2, 3}, []int{1, 2}, "optional assertion explanation")
	assert.Must(tb).Contain(tb,
		map[string]int{"The Answer": 42, "oth": 13},
		map[string]int{"The Answer": 42},
		"optional assertion explanation")
}

func ExampleNotContain() {
	var tb testing.TB
	assert.Must(tb).NotContain(tb, []int{1, 2, 3}, 42)
	assert.Must(tb).NotContain(tb, []int{1, 2, 3}, []int{1, 2, 42})
	assert.Must(tb).NotContain(tb,
		map[string]int{"The Answer": 42, "oth": 13},
		map[string]int{"The Answer": 41})
}

func ExampleContainExactly() {
	var tb testing.TB
	assert.ContainExactly(tb, []int{1, 2, 3}, []int{2, 3, 1}, "optional assertion explanation")  // true
	assert.ContainExactly(tb, []int{1, 2, 3}, []int{1, 42, 2}, "optional assertion explanation") // false
}

func ExampleErrorIs() {
	var tb testing.TB
	actualErr := errors.New("boom")
	assert.ErrorIs(tb, errors.New("boom"), actualErr)                                  // passes for equality
	assert.ErrorIs(tb, errors.New("boom"), fmt.Errorf("wrapped error: %w", actualErr)) // passes for wrapped errors
}
