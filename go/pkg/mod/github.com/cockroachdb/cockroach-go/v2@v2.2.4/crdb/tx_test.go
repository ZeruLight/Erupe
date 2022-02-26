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

package crdb

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/cockroachdb/cockroach-go/v2/testserver"
)

// TestExecuteTx verifies transaction retry using the classic
// example of write skew in bank account balance transfers.
func TestExecuteTx(t *testing.T) {
	db, stop := testserver.NewDBForTest(t)
	defer stop()
	ctx := context.Background()

	if err := ExecuteTxGenericTest(ctx, stdlibWriteSkewTest{db: db}); err != nil {
		t.Fatal(err)
	}
}

type stdlibWriteSkewTest struct {
	db *sql.DB
}

var _ WriteSkewTest = stdlibWriteSkewTest{}

func (t stdlibWriteSkewTest) Init(ctx context.Context) error {
	initStmt := `
CREATE DATABASE d;
CREATE TABLE d.t (acct INT PRIMARY KEY, balance INT);
INSERT INTO d.t (acct, balance) VALUES (1, 100), (2, 100);
`
	_, err := t.db.ExecContext(ctx, initStmt)
	return err
}

func (t stdlibWriteSkewTest) ExecuteTx(ctx context.Context, fn func(tx interface{}) error) error {
	return ExecuteTx(ctx, t.db, nil /* opts */, func(tx *sql.Tx) error {
		return fn(tx)
	})
}

func (t stdlibWriteSkewTest) GetBalances(ctx context.Context, txi interface{}) (int, int, error) {
	tx := txi.(*sql.Tx)
	var rows *sql.Rows
	rows, err := tx.QueryContext(ctx, `SELECT balance FROM d.t WHERE acct IN (1, 2);`)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	var bal1, bal2 int
	balances := []*int{&bal1, &bal2}
	i := 0
	for ; rows.Next(); i++ {
		if err = rows.Scan(balances[i]); err != nil {
			return 0, 0, err
		}
	}
	if i != 2 {
		return 0, 0, fmt.Errorf("expected two balances; got %d", i)
	}
	return bal1, bal2, nil
}

func (t stdlibWriteSkewTest) UpdateBalance(
	ctx context.Context, txi interface{}, acct, delta int,
) error {
	tx := txi.(*sql.Tx)
	_, err := tx.ExecContext(ctx, `UPDATE d.t SET balance=balance+$1 WHERE acct=$2;`, delta, acct)
	return err
}
