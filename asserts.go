package testcase

import (
	"testing"

	"github.com/adamluzsi/testcase/assert"
)

func makeIt(tb testing.TB) It {
	return It{
		Must:   assert.Must(tb),
		Should: assert.Should(tb),
	}
}

type It struct {
	// Must Asserter will use FailNow on a failed assertion.
	// This will make test exit early on.
	Must Asserter
	// Should Asserter's will allow to continue the test scenario,
	// but mark test failed on a failed assertion.
	Should Asserter
}

// Asserter contains a minimum set of assertion interactions.
type Asserter interface {
	True(v bool, msg ...interface{})
	False(v bool, msg ...interface{})
	Nil(v interface{}, msg ...interface{})
	NotNil(v interface{}, msg ...interface{})
	Equal(expected, actually interface{}, msg ...interface{})
	NotEqual(expected, actually interface{}, msg ...interface{})
	Contain(source, sub interface{}, msg ...interface{})
	NotContain(source, sub interface{}, msg ...interface{})
	ContainExactly(expected, actually interface{}, msg ...interface{})
	Panic(blk func(), msg ...interface{}) (panicValue interface{})
	NotPanic(blk func(), msg ...interface{})
	Empty(v interface{}, msg ...interface{})
	NotEmpty(v interface{}, msg ...interface{})
}
