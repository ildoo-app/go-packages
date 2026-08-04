package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	Begin(context.Context) (pgx.Tx, error)
	Commit(context.Context) error
	Rollback(context.Context) error
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
