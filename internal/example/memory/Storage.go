package memory

import (
	"context"
)

// example factory
func NewStorage() *Storage {
	//sql.Open(`driver`, connstr) ...
	return &Storage{}
}

type Storage struct {
	table map[string]any
}

func (p Storage) Close() error {
	panic("implement me")
}

func (p Storage) BeginTx(ctx context.Context) (context.Context, error) {
	panic("implement me")
}

func (p Storage) CommitTx(ctx context.Context) error {
	panic("implement me")
}

func (p Storage) RollbackTx(ctx context.Context) error {
	panic("implement me")
}
