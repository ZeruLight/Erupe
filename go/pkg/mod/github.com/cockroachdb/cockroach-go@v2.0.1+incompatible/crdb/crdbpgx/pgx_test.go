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

package crdbpgx

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/cockroachdb/cockroach-go/testserver"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgxpool"
	"testing"
)

// TestExecuteTx verifies transaction retry using the classic
// example of write skew in bank account balance transfers.
func TestExecuteTx(t *testing.T) {
	ts, err := testserver.NewTestServer()
	if err != nil {
		t.Fatal(err)
	}
	if err := ts.Start(); err != nil {
		t.Fatal(err)
	}
	url := ts.PGURL()
	db, err := sql.Open("postgres", url.String())
	if err != nil {
		t.Fatal(err)
	}
	if err := ts.WaitForInit(db); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	pool, err := pgxpool.Connect(ctx, ts.PGURL().String())
	if err != nil {
		t.Fatal(err)
	}

	if err := crdb.ExecuteTxGenericTest(ctx, pgxWriteSkewTest{pool: pool}); err != nil {
		t.Fatal(err)
	}
}

type pgxWriteSkewTest struct {
	pool *pgxpool.Pool
}

func (t pgxWriteSkewTest) Init(ctx context.Context) error {
	initStmt := `
CREATE DATABASE d;
CREATE TABLE d.t (acct INT PRIMARY KEY, balance INT);
INSERT INTO d.t (acct, balance) VALUES (1, 100), (2, 100);
`
	_, err := t.pool.Exec(ctx, initStmt)
	return err
}

var _ crdb.WriteSkewTest = pgxWriteSkewTest{}

// ExecuteTx is part of the crdb.WriteSkewTest interface.
func (t pgxWriteSkewTest) ExecuteTx(ctx context.Context, fn func(tx interface{}) error) error {
	return ExecuteTx(ctx, t.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		return fn(tx)
	})
}

// GetBalances is part of the crdb.WriteSkewTest interface.
func (t pgxWriteSkewTest) GetBalances(ctx context.Context, txi interface{}) (int, int, error) {
	tx := txi.(pgx.Tx)
	var rows pgx.Rows
	rows, err := tx.Query(ctx, `SELECT balance FROM d.t WHERE acct IN (1, 2);`)
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

// UpdateBalance is part of the crdb.WriteSkewInterface.
func (t pgxWriteSkewTest) UpdateBalance(
	ctx context.Context, txi interface{}, acct, delta int,
) error {
	tx := txi.(pgx.Tx)
	_, err := tx.Exec(ctx, `UPDATE d.t SET balance=balance+$1 WHERE acct=$2;`, delta, acct)
	return err
}
