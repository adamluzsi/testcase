package mydomain

import "context"

type MyUseCaseInteractor struct {
	Storage // example dependency
}

type Storage interface {
}

func (i *MyUseCaseInteractor) Foo(ctx context.Context) (string, error) {
	return `bar`, nil
}
