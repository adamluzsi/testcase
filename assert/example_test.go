package assert_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"go.llib.dev/testcase/random"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
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

func ExampleSub() {
	var tb testing.TB
	assert.Sub(tb, []int{1, 2, 3}, []int{1, 2}, "optional assertion explanation")
}

func ExampleAsserter_Sub() {
	var tb testing.TB
	assert.Must(tb).Sub([]int{1, 2, 3}, 3, "optional assertion explanation")
	assert.Must(tb).Sub([]int{1, 2, 3}, []int{1, 2}, "optional assertion explanation")
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
	assert.Must(tb).AnyOf(func(anyOf *assert.A) {
		for _, testingCase := range list {
			anyOf.Case(func(it assert.It) {
				it.Must.True(testingCase.Bar())
			})
		}
	})
}

func ExampleAnyOf_anyOfTheElement() {
	var tb testing.TB
	var list []interface {
		Foo() int
		Bar() bool
		Baz() string
	}
	assert.AnyOf(tb, func(anyOf *assert.A) {
		for _, testingCase := range list {
			anyOf.Case(func(it assert.It) {
				it.Must.True(testingCase.Bar())
			})
		}
	})
}

func ExampleAnyOf_anyOfExpectedOutcome() {
	var tb testing.TB
	var rnd = random.New(random.CryptoSeed{})

	outcome := rnd.Bool()

	assert.AnyOf(tb, func(a *assert.A) {
		a.Case(func(it assert.It) {
			it.Must.True(outcome)
		})

		a.Case(func(it assert.It) {
			it.Must.False(outcome)
		})
	})
}

func ExampleAnyOf_listOfInterface() {
	var tb testing.TB
	type ExampleInterface interface {
		Foo() int
		Bar() bool
		Baz() string
	}
	anyOf := assert.A{TB: tb, Fail: tb.FailNow}
	for _, v := range []ExampleInterface{} {
		anyOf.Case(func(it assert.It) {
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
	anyOf := assert.A{TB: tb, Fail: tb.FailNow}
	for _, v := range []BigStruct{} {
		anyOf.Case(func(it assert.It) {
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
	anyOf := assert.A{TB: tb, Fail: tb.FailNow}
	for _, v := range []StructWithDynamicValues{} {
		anyOf.Case(func(it assert.It) {
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
	anyOf := assert.A{TB: tb, Fail: tb.FailNow}
	anyOf.Case(func(it assert.It) {
		it.Must.Equal(`foo`, es.Type)
		it.Must.Equal(1, es.A)
		it.Must.Equal(2, es.B)
		it.Must.Equal(3, es.C)
	})
	anyOf.Case(func(it assert.It) {
		it.Must.Equal(`foo`, es.Type)
		it.Must.Equal(3, es.A)
		it.Must.Equal(2, es.B)
		it.Must.Equal(1, es.C)
	})
	anyOf.Case(func(it assert.It) {
		it.Must.Equal(`bar`, es.Type)
		it.Must.Equal(11, es.A)
		it.Must.Equal(12, es.B)
		it.Must.Equal(13, es.C)
	})
	anyOf.Case(func(it assert.It) {
		it.Must.Equal(`baz`, es.Type)
		it.Must.Equal(21, es.A)
		it.Must.Equal(22, es.B)
		it.Must.Equal(23, es.C)
	})
	anyOf.Finish()
}

type ExamplePublisherEvent struct{ V int }
type ExamplePublisher struct{}

func (ExamplePublisher) Publish(ExamplePublisherEvent)         {}
func (ExamplePublisher) Subscribe(func(ExamplePublisherEvent)) {}
func (ExamplePublisher) Wait()                                 {}
func (ExamplePublisher) Close() error                          { return nil }

func ExampleAnyOf_fanOutPublishing() {
	var tb testing.TB
	publisher := ExamplePublisher{}
	anyOf := &assert.A{TB: tb, Fail: tb.FailNow}
	for i := 0; i < 42; i++ {
		publisher.Subscribe(func(event ExamplePublisherEvent) {
			anyOf.Case(func(it assert.It) {
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

type ExampleEqualableWithIsEqual struct {
	IrrelevantExportedField int
	relevantUnexportedValue int
}

func (es ExampleEqualableWithIsEqual) IsEqual(oth ExampleEqualableWithIsEqual) bool {
	return es.relevantUnexportedValue == oth.relevantUnexportedValue
}

type ExampleEqualableWithEqual struct {
	IrrelevantExportedField int
	relevantUnexportedValue int
}

func (es ExampleEqualableWithEqual) IsEqual(oth ExampleEqualableWithEqual) bool {
	return es.relevantUnexportedValue == oth.relevantUnexportedValue
}

func ExampleAsserter_Equal_withIsEqualMethod() {
	var tb testing.TB

	expected := ExampleEqualableWithIsEqual{
		IrrelevantExportedField: 42,
		relevantUnexportedValue: 24,
	}

	actual := ExampleEqualableWithIsEqual{
		IrrelevantExportedField: 4242,
		relevantUnexportedValue: 24,
	}

	assert.Must(tb).Equal(expected, actual) // passes as by IsEqual terms the two value is equal
}

func ExampleAsserter_Equal_withEqualMethod() {
	var tb testing.TB

	expected := ExampleEqualableWithEqual{
		IrrelevantExportedField: 42,
		relevantUnexportedValue: 24,
	}

	actual := ExampleEqualableWithEqual{
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
	assert.Must(tb).Contain([]int{1, 2, 3}, 3, "optional assertion explanation")
	assert.Must(tb).Contain([]int{1, 2, 3}, []int{1, 2}, "optional assertion explanation")
	assert.Must(tb).Contain(
		map[string]int{"The Answer": 42, "oth": 13},
		map[string]int{"The Answer": 42},
		"optional assertion explanation")
}

func ExampleNotContain() {
	var tb testing.TB
	assert.Must(tb).NotContain([]int{1, 2, 3}, 42)
	assert.Must(tb).NotContain([]int{1, 2, 3}, []int{1, 2, 42})
	assert.Must(tb).NotContain(
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
	assert.ErrorIs(tb, actualErr, errors.New("boom"))                                  // passes for equality
	assert.ErrorIs(tb, fmt.Errorf("wrapped error: %w", actualErr), errors.New("boom")) // passes for wrapped errors
}

func ExampleWaiter_Wait() {
	w := assert.Waiter{WaitDuration: time.Millisecond}

	w.Wait() // will wait 1 millisecond and attempt to schedule other go routines
}

func ExampleWaiter_While() {
	w := assert.Waiter{
		WaitDuration: time.Millisecond,
		Timeout:      time.Second,
	}

	// will attempt to wait until condition returns false.
	// The maximum time it is willing to wait is equal to the wait timeout duration.
	w.While(func() bool {
		return rand.Intn(1) == 0
	})
}

func ExampleMakeRetry() {
	var tb testing.TB
	assert.MakeRetry(5*time.Second).Assert(tb, func(it assert.It) {
		// use "it" as you would tb, but if the test fails with "it"
		// then the function block will be retried until the allowed time duration, which is one minute in this case.
	})
}

func ExampleMakeRetry_byCount() {
	var tb testing.TB
	assert.MakeRetry(3 /* times */).Assert(tb, func(it assert.It) {
		// use "it" as you would tb, but if the test fails with "it"
		// it will be retried 3 times as specified above as argument.
	})
}

func ExampleMakeRetry_byTimeout() {
	var tb testing.TB
	assert.MakeRetry(time.Minute /* times */).Assert(tb, func(it assert.It) {
		// use "it" as you would tb, but if the test fails with "it"
		// then the function block will be retried until the allowed time duration, which is one minute in this case.
	})
}

func ExampleRetry() {
	waiter := assert.Waiter{
		WaitDuration: time.Millisecond,
		Timeout:      time.Second,
	}
	w := assert.Retry{Strategy: waiter}

	var t *testing.T
	// will attempt to wait until assertion block passes without a failing testCase result.
	// The maximum time it is willing to wait is equal to the wait timeout duration.
	// If the wait timeout reached, and there was no passing assertion run,
	// the last failed assertion history is replied to the received testing.TB
	//   In this case the failure would be replied to the *testing.T.
	w.Assert(t, func(it assert.It) {
		if rand.Intn(1) == 0 {
			it.Fatal(`boom`)
		}
	})
}

func ExampleRetry_asContextOption() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test(`flaky`, func(t *testcase.T) {
		// flaky test content here
	}, testcase.Flaky(assert.RetryCount(42)))
}

func ExampleRetry_count() {
	_ = assert.Retry{Strategy: assert.RetryCount(42)}
}

func ExampleRetry_byTimeout() {
	r := assert.Retry{Strategy: assert.Waiter{
		WaitDuration: time.Millisecond,
		Timeout:      time.Second,
	}}

	var t *testing.T
	r.Assert(t, func(it assert.It) {
		if rand.Intn(1) == 0 {
			it.Fatal(`boom`)
		}
	})
}

func ExampleRetry_byCount() {
	r := assert.Retry{Strategy: assert.RetryCount(42)}

	var t *testing.T
	r.Assert(t, func(it assert.It) {
		if rand.Intn(1) == 0 {
			it.Fatal(`boom`)
		}
	})
}

func ExampleRetry_byCustomRetryStrategy() {
	// this approach ideal if you need to deal with asynchronous systems
	// where you know that if a workflow process ended already,
	// there is no point in retrying anymore the assertion.

	while := func(isFailed func() bool) {
		for isFailed() {
			// just retry while assertion is failed
			// could be that assertion will be failed forever.
			// Make sure the assertion is not stuck in a infinite loop.
		}
	}

	r := assert.Retry{Strategy: assert.RetryStrategyFunc(while)}

	var t *testing.T
	r.Assert(t, func(it assert.It) {
		if rand.Intn(1) == 0 {
			it.Fatal(`boom`)
		}
	})
}

func ExampleAsserter_Error() {
	var tb testing.TB
	asserter := assert.Should(tb)
	asserter.Error(nil)                // fail
	asserter.Error(errors.New("boom")) // pass
}

func ExampleError() {
	var tb testing.TB
	assert.Error(tb, nil)                // fail
	assert.Error(tb, errors.New("boom")) // pass
}

func ExampleAsserter_NoError() {
	var tb testing.TB
	asserter := assert.Should(tb)
	asserter.NoError(nil)                // pass
	asserter.NoError(errors.New("boom")) // fail
}

func ExampleNoError() {
	var tb testing.TB
	assert.NoError(tb, nil)                // pass
	assert.NoError(tb, errors.New("boom")) // fail
}

func ExampleAsserter_Read() {
	var tb testing.TB
	must := assert.Must(tb)
	must.Read("expected content", strings.NewReader("expected content"))  // pass
	must.Read("expected content", strings.NewReader("different content")) // fail
}

func ExampleRead() {
	var tb testing.TB
	assert.Read(tb, "expected content", strings.NewReader("expected content"))  // pass
	assert.Read(tb, "expected content", strings.NewReader("different content")) // fail
}

func ExampleAsserter_ReadAll() {
	var tb testing.TB
	must := assert.Must(tb)
	content := must.ReadAll(strings.NewReader("expected content")) // pass
	_ = content
	must.ReadAll(iotest.ErrReader(errors.New("boom"))) // fail
}

func ExampleReadAll() {
	var tb testing.TB
	content := assert.ReadAll(tb, strings.NewReader("expected content")) // pass
	_ = content
	assert.ReadAll(tb, iotest.ErrReader(errors.New("boom"))) // fail
}

func Example_configureDiffFunc() {
	assert.DiffFunc = func(value, othValue any) string {
		return fmt.Sprintf("%#v | %#v", value, othValue)
	}

	var tb testing.TB
	assert.Equal(tb, "foo", "bar")
}

func ExampleWithin() {
	var tb testing.TB

	assert.Within(tb, time.Second, func(ctx context.Context) {
		// OK
	})

	assert.Within(tb, time.Nanosecond, func(ctx context.Context) {
		time.Sleep(time.Second)
		// FAIL
	})
}

func ExampleAsserter_Within() {
	var tb testing.TB
	a := assert.Must(tb)

	a.Within(time.Second, func(ctx context.Context) {
		// OK
	})

	a.Within(time.Nanosecond, func(ctx context.Context) {
		time.Sleep(time.Second)
		// FAIL
	})
}

func ExampleNotWithin() {
	var tb testing.TB

	assert.NotWithin(tb, time.Second, func(ctx context.Context) {
		return // FAIL
	})

	assert.NotWithin(tb, time.Nanosecond, func(ctx context.Context) {
		time.Sleep(time.Second) // OK
	})
}

func ExampleAsserter_NotWithin() {
	var tb testing.TB
	a := assert.Must(tb)

	a.NotWithin(time.Second, func(ctx context.Context) {
		return // FAIL
	})

	a.NotWithin(time.Nanosecond, func(ctx context.Context) {
		time.Sleep(time.Second) // OK
	})
}

func ExampleOneOf() {
	var tb testing.TB
	values := []string{"foo", "bar", "baz"}

	assert.OneOf(tb, values, func(it assert.It, got string) {
		it.Must.Equal("bar", got)
	}, "optional assertion explanation")
}

func ExampleAsserter_OneOf() {
	var tb testing.TB
	values := []string{"foo", "bar", "baz"}

	assert.Must(tb).OneOf(values, func(it assert.It, got string) {
		it.Must.Equal("bar", got)
	}, "optional assertion explanation")
}

func ExampleMatchRegexp() {
	var tb testing.TB
	assert.MatchRegexp(tb, "42", "[0-9]+")
	assert.MatchRegexp(tb, "forty-two", "[a-z]+")
	assert.MatchRegexp(tb, []byte("forty-two"), "[a-z]+")
}

func ExampleAsserter_MatchRegexp() {
	var tb testing.TB
	assert.Must(tb).MatchRegexp("42", "[0-9]+")
	assert.Must(tb).MatchRegexp("forty-two", "[a-z]+")
}

func ExampleNotMatchRegexp() {
	var tb testing.TB
	assert.NotMatchRegexp(tb, "42", "^[a-z]+")
	assert.NotMatchRegexp(tb, "forty-two", "^[0-9]+")
	assert.NotMatchRegexp(tb, []byte("forty-two"), "^[0-9]+")
}

func ExampleAsserter_NotMatchRegexp() {
	var tb testing.TB
	assert.Must(tb).NotMatchRegexp("42", "^[a-z]+")
	assert.Must(tb).NotMatchRegexp("forty-two", "^[0-9]+")
}

func ExampleAsserter_Eventually() {
	var tb testing.TB
	assert.Must(tb).Eventually(time.Minute, func(it assert.It) {
		it.Must.True(rand.Intn(1) == 0)
	})
}

func ExampleEventually() {
	var tb testing.TB
	assert.Eventually(tb, time.Second, func(it assert.It) {
		it.Must.True(rand.Intn(1) == 0)
	})
}

func ExampleAsserter_Unique() {
	var tb testing.TB
	assert.Must(tb).Unique([]int{1, 2, 3}, "expected of unique values")
}

func ExampleUnique() {
	var tb testing.TB
	assert.Unique(tb, []int{1, 2, 3}, "expected of unique values")
}

func ExampleAsserter_NotUnique() {
	var tb testing.TB
	assert.Must(tb).NotUnique([]int{1, 2, 3, 1}, "expected of a list with at least one duplicate")
}

func ExampleNotUnique() {
	var tb testing.TB
	assert.NotUnique(tb, []int{1, 2, 3, 1}, "expected of a list with at least one duplicate")
}
