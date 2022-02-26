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

package crdbgorm

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"gorm.io/gorm"
)

// ExecuteTx runs fn inside a transaction and retries it as needed. On
// non-retryable failures, the transaction is aborted and rolled back; on
// success, the transaction is committed.
//
// See crdb.ExecuteTx() for more information.
func ExecuteTx(
	ctx context.Context, db *gorm.DB, opts *sql.TxOptions, fn func(tx *gorm.DB) error,
) error {
	tx := db.WithContext(ctx).Begin(opts)
	if db.Error != nil {
		return db.Error
	}
	return crdb.ExecuteInTx(ctx, gormTxAdapter{tx}, func() error { return fn(tx) })
}

// gormTxAdapter adapts a *gorm.DB to a crdb.Tx.
type gormTxAdapter struct {
	db *gorm.DB
}

var _ crdb.Tx = gormTxAdapter{}

// Exec is part of the crdb.Tx interface.
func (tx gormTxAdapter) Exec(_ context.Context, q string, args ...interface{}) error {
	return tx.db.Exec(q, args...).Error
}

// Commit is part of the crdb.Tx interface.
func (tx gormTxAdapter) Commit(_ context.Context) error {
	return tx.db.Commit().Error
}

// Rollback is part of the crdb.Tx interface.
func (tx gormTxAdapter) Rollback(_ context.Context) error {
	return tx.db.Rollback().Error
}
