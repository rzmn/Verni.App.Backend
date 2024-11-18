package db_mock

import (
	"context"
	"database/sql"
)

type DbMock struct {
	QueryImpl    func(query string, args ...any) (*sql.Rows, error)
	QueryRowImpl func(query string, args ...any) *sql.Row
	ExecImpl     func(query string, args ...any) (sql.Result, error)
	BeginTxImpl  func(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	PrepareImpl  func(query string) (*sql.Stmt, error)
	CloseImpl    func() error
}

func (c *DbMock) Query(query string, args ...any) (*sql.Rows, error) {
	return c.QueryImpl(query, args...)
}

func (c *DbMock) QueryRow(query string, args ...any) *sql.Row {
	return c.QueryRowImpl(query, args...)
}

func (c *DbMock) Exec(query string, args ...any) (sql.Result, error) {
	return c.ExecImpl(query, args...)
}

func (c *DbMock) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return c.BeginTxImpl(ctx, opts)
}

func (c *DbMock) Prepare(query string) (*sql.Stmt, error) {
	return c.PrepareImpl(query)
}

func (c *DbMock) Close() error {
	return c.CloseImpl()
}
