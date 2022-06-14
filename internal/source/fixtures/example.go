package fixtures

import "context"

type MyType struct {
	IntField int
}

func (MyType) MyFunc(ctx context.Context) error {
	return nil
}
