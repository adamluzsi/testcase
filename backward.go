package testcase

import "go.llib.dev/testcase/assert"

// Let is a method to provide backward compatibility with the existing testing suite.
// Due to how Go type parameters work, methods are not allowed to have type parameters,
// thus Let has moved to be a pkg-level function in the package.
//
// DEPRECATED: use testcase.Let instead testcase#Spec.Let.
func (spec *Spec) Let(varName VarID, blk VarInit[any]) Var[any] {
	return let[any](spec, varName, blk)
}

// LetValue is a method to provide backward compatibility with the existing testing suite.
// Due to how Go type parameters work, methods are not allowed to have type parameters,
// thus LetValue has moved to be a pkg-level function in the package.
//
// DEPRECATED: use testcase.LetValue instead testcase#Spec.LetValue.
func (spec *Spec) LetValue(varName VarID, value any) Var[any] {
	return letValue[any](spec, varName, value)
}

// VarInitFunc is a backward compatibility type for VarInit.
//
// DEPRECATED: use VarInit type instead.
type VarInitFunc[V any] func(*T) V

// RetryStrategyForEventually
//
// DEPRECATED: use testcase.WithRetryStrategy instead
func RetryStrategyForEventually(strategy assert.RetryStrategy) SpecOption {
	return WithRetryStrategy(strategy)
}
