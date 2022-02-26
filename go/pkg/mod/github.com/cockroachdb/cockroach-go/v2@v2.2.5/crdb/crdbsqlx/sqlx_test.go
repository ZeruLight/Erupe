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
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
)

// TestExecuteTx verifies transaction retry using the classic
// example of write skew in bank account balance transfers.
func TestExecuteTx(t *testing.T) {
	ts, err := testserver.NewTestServer()
	if err != nil {
		t.Fatal(err)
	}
	url := ts.PGURL()
	db, err := sqlx.Connect("postgres", url.String())
	if err != nil {
		t.Fatal(err)
	}
	if err := ts.WaitForInit(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := crdb.ExecuteTxGenericTest(ctx, sqlxConnSkewTest{db: db}); err != nil {
		t.Fatal(err)
	}
}

type sqlxConnSkewTest struct {
	db *sqlx.DB
}

func (t sqlxConnSkewTest) Init(ctx context.Context) error {
	initStmt := `
CREATE DATABASE d;
CREATE TABLE d.t (acct INT PRIMARY KEY, balance INT);
INSERT INTO d.t (acct, balance) VALUES (1, 100), (2, 100);
`
	_, err := t.db.ExecContext(ctx, initStmt)
	return err
}

var _ crdb.WriteSkewTest = sqlxConnSkewTest{}

// ExecuteTx is part of the crdb.WriteSkewTest interface.
func (t sqlxConnSkewTest) ExecuteTx(ctx context.Context, fn func(tx interface{}) error) error {
	return ExecuteTx(ctx, t.db, nil /* txOptions */, func(tx *sqlx.Tx) error {
		return fn(tx)
	})
}

// GetBalances is part of the crdb.WriteSkewTest interface.
func (t sqlxConnSkewTest) GetBalances(ctx context.Context, txi interface{}) (int, int, error) {
	tx := txi.(*sqlx.Tx)
	rows, err := tx.QueryContext(ctx, `SELECT balance FROM d.t WHERE acct IN (1, 2);`)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	balances := []int{}
	for rows.Next() {
		var bal int
		if err = rows.Scan(&bal); err != nil {
			return 0, 0, err
		}
		balances = append(balances, bal)
	}
	if len(balances) != 2 {
		return 0, 0, fmt.Errorf("expected two balances; got %d", len(balances))
	}
	return balances[0], balances[1], nil
}

// UpdateBalance is part of the crdb.WriteSkewInterface.
func (t sqlxConnSkewTest) UpdateBalance(ctx context.Context, txi interface{}, acct, delta int) error {
	tx := txi.(*sqlx.Tx)
	_, err := tx.ExecContext(ctx, `UPDATE d.t SET balance=balance+$1 WHERE acct=$2;`, delta, acct)
	return err
}
