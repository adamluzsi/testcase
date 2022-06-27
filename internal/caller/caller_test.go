package caller_test

import (
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/caller"
)

func TestGetCall(t *testing.T) {
	t.Log("from function's top level")
	cfn, ok := caller.GetFunc()
	assert.True(t, ok)
	assert.Equal(t, "TestGetCall", cfn.Funcion)
	assert.Equal(t, "", cfn.Receiver)
	assert.Equal(t, "caller_test", cfn.Package)

	t.Log("from lambda that is part of a function")
	func() {
		cfn, ok := caller.GetFunc()
		assert.True(t, ok)
		assert.Equal(t, "TestGetCall", cfn.Funcion)
		assert.Equal(t, "", cfn.Receiver)
		assert.Equal(t, "caller_test", cfn.Package)
	}()

	getCallFixture := GetCallFixture{}

	t.Log("from the top of a method on a receiver")
	getCallFixture.TestMethod(t)

	t.Log("from the top of a method on a pointer receiver")
	getCallFixture.TestPointerMethod(t)

	t.Log("from a lambda inside a method on a receiver")
	getCallFixture.TestLambdaInMethod(t)
}

type GetCallFixture struct{}

func (GetCallFixture) TestMethod(tb testing.TB) {
	cfn, ok := caller.GetFunc()
	assert.True(tb, ok)
	assert.Equal(tb, "TestMethod", cfn.Funcion)
	assert.Equal(tb, "GetCallFixture", cfn.Receiver)
	assert.Equal(tb, "caller_test", cfn.Package)
}
func (*GetCallFixture) TestPointerMethod(tb testing.TB) {
	cfn, ok := caller.GetFunc()
	assert.True(tb, ok)
	assert.Equal(tb, "TestPointerMethod", cfn.Funcion)
	assert.Equal(tb, "*GetCallFixture", cfn.Receiver)
	assert.Equal(tb, "caller_test", cfn.Package)
}

func (f GetCallFixture) TestLambdaInMethod(tb testing.TB) {
	var run bool
	func() {
		func() {
			cfn, ok := caller.GetFunc()
			assert.True(tb, ok)
			assert.Equal(tb, "TestLambdaInMethod", cfn.Funcion)
			assert.Equal(tb, "GetCallFixture", cfn.Receiver)
			assert.Equal(tb, "caller_test", cfn.Package)
			run = true
		}()
	}()
	assert.True(tb, run)
}
