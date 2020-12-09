package mydomain

import "context"

type MyUseCaseInteractor struct {
	Storage // example dependency
}

type Storage interface {
	BeginTx(context.Context) (context.Context, error)
	CommitTx(context.Context) error
	RollbackTx(context.Context) error
}

func (i *MyUseCaseInteractor) Foo(ctx context.Context) (string, error) {
	return `bar`, nil
}
