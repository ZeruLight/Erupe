package crdbsqlx

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/cockroachdb/cockroach-go/crdb"
)

// ExecuteTx runs fn inside a transaction and retries it as needed. On
// non-retryable failures, the transaction is aborted and rolled back; on
// success, the transaction is committed.
//
// See crdb.ExecuteTx() for more information.
func ExecuteTx(ctx context.Context, db *sqlx.DB, opts *sql.TxOptions, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return err
	}
	return crdb.ExecuteInTx(ctx, sqlxTxAdapter{tx}, func() error { return fn(tx) })
}

// sqlxTxAdapter adapts a *sqlx.tx to a crdb.Tx.
type sqlxTxAdapter struct {
	*sqlx.Tx
}

var _ crdb.Tx = sqlxTxAdapter{}

func (s sqlxTxAdapter) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := s.Tx.ExecContext(ctx, query, args...)
	return err
}

func (s sqlxTxAdapter) Commit(ctx context.Context) error {
	return s.Tx.Commit()
}

func (s sqlxTxAdapter) Rollback(ctx context.Context) error {
	return s.Tx.Rollback()
}
