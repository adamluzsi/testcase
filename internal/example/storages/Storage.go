package storages

import (
	"context"
	"database/sql"
)

// example factory
func New(connstr string) (*Storage, error) {
	//sql.Open(`driver`, connstr) ...
	return &Storage{}, nil
}

type Storage struct {
	DB *sql.DB
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
