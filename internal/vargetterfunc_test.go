package internal_test

import (
	"go.llib.dev/testcase"
	"go.llib.dev/testcase/internal"
	"go.llib.dev/testcase/internal/testent"
)

var _ testcase.VarGetter[testent.Foo] = internal.VarGetterFunc[testcase.T, testent.Foo](nil)
