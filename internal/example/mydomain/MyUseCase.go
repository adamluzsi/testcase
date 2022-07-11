package mydomain

import (
	"context"
	"strings"
)

type MyUseCase struct {
	Storage // example dependency
}

type Storage interface {
	BeginTx(context.Context) (context.Context, error)
	CommitTx(context.Context) error
	RollbackTx(context.Context) error
}

func (i *MyUseCase) MyFunc() {}

func (i *MyUseCase) MyFuncThatNeedsSomething(something any) {}

func (i *MyUseCase) Foo(ctx context.Context) (string, error) {
	return `bar`, nil
}

func (i *MyUseCase) IsLower(s string) bool {
	return strings.ToLower(s) == s
}

func (i *MyUseCase) ThreadSafeCall() {}
