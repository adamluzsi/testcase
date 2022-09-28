package testcase

import "github.com/adamluzsi/testcase/assert"

type (
	// Eventually
	//
	// DEPRECATED: use assert.Eventually instead
	Eventually = assert.Eventually
	// RetryStrategy
	//
	// DEPRECATED: use assert.RetryStrategy instead
	RetryStrategy = assert.RetryStrategy
	// RetryStrategyFunc
	//
	// DEPRECATED: use assert.RetryStrategyFunc instead
	RetryStrategyFunc = assert.RetryStrategyFunc
	// Waiter
	//
	// DEPRECATED: use assert.Waiter instead
	Waiter = assert.Waiter
)

// RetryCount is moved from this package.
//
// DEPRECATED: use assert.RetryCount instead
func RetryCount(times int) assert.RetryStrategy {
	return assert.RetryCount(times)
}

// Let is a method to provide backward compatibility with the existing testing suite.
// Due to how Go type parameters work, methods are not allowed to have type parameters,
// thus Let has moved to be a pkg-level function in the package.
//
// DEPRECATED: use testcase.Let instead testcase#Spec.Let.
func (spec *Spec) Let(varName string, blk VarInit[any]) Var[any] {
	return let[any](spec, varName, blk)
}

// LetValue is a method to provide backward compatibility with the existing testing suite.
// Due to how Go type parameters work, methods are not allowed to have type parameters,
// thus LetValue has moved to be a pkg-level function in the package.
//
// DEPRECATED: use testcase.LetValue instead testcase#Spec.LetValue.
func (spec *Spec) LetValue(varName string, value any) Var[any] {
	return letValue[any](spec, varName, value)
}

// VarInitFunc is a backward compatibility type for VarInit.
//
// DEPRECATED: use VarInit type instead.
type VarInitFunc[V any] func(*T) V
