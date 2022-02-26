// Copyright 2016 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package crdbsqlx

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
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
