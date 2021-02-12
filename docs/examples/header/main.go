package header

import (
	"context"
	"database/sql"
	"database/sql/driver"
)

// package mystorage

// sqlDB is the header interface to *sql.DB
type sqlDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Ping() error
	Driver() driver.Driver
	Close() error
	Conn(ctx context.Context) (*sql.Conn, error)
	PingContext(ctx context.Context) error
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	// ... and so on
}

type Supplier struct {
	db sqlDB
}

func (m Supplier) Count() error {
	var count int
	if err := m.db.QueryRow(`SELECT 1`).Scan(&count); err != nil {
		return err
	}

	return nil
}
