package expects

import (
	"github.com/adamluzsi/testcase/assert"
	"testing"
)

type ValueExpectation[T any] interface{ To(Assertion[T]) }

func Expect[T any](tb testing.TB, expected T) ValueExpectation[T] {
	return expectation[T]{TB: tb, Expected: expected}
}

type expectation[T any] struct {
	TB testing.TB
	
	Expected T
}

func (e expectation[T]) To(a Assertion[T]) {
	a.Assert(e.TB, e.Expected)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Assertion[T any] interface {
	Assert(tb testing.TB, expected T)
}

type assertFunc[T any] func(testing.TB, T)

func (fn assertFunc[T]) Assert(tb testing.TB, expected T) { fn(tb, expected) }

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func Equal[T any](got T) Assertion[T] {
	return assertFunc[T](func(tb testing.TB, expected T) {
		assert.Equal(tb, expected, got)
	})
}
