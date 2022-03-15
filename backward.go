package testcase

// Let is a method to provide backward compatibility with the existing testing suite.
// Due to how Go type parameters work, methods are not allowed to have type parameters,
// thus Let has moved to be a pkg-level function in the package.
//
// DEPRECATED: use testcase.Let instead testcase#Spec.Let.
func (spec *Spec) Let(varName string, blk varInitBlk[any]) Var[any] {
	return let[any](spec, varName, blk)
}

// LetValue is a method to provide backward compatibility with the existing testing suite.
// Due to how Go type parameters work, methods are not allowed to have type parameters,
// thus LetValue has moved to be a pkg-level function in the package.
// DEPRECATED: use testcase.LetValue instead testcase#Spec.LetValue.
func (spec *Spec) LetValue(varName string, value any) Var[any] {
	return letValue[any](spec, varName, value)
}
