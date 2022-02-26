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
	"fmt"
	"testing"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestExecuteTx verifies transaction retry using the classic example of write
// skew in bank account balance transfers.
func TestExecuteTx(t *testing.T) {
	ts, err := testserver.NewTestServer()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	gormDB, err := gorm.Open(postgres.Open(ts.PGURL().String()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// Set to logger.Info and gorm logs all the queries.
	gormDB.Logger.LogMode(logger.Silent)

	if err := crdb.ExecuteTxGenericTest(ctx, gormWriteSkewTest{db: gormDB}); err != nil {
		t.Fatal(err)
	}
}

// Account is our model, which corresponds to the "accounts" database
// table.
type Account struct {
	ID      int `gorm:"primary_key"`
	Balance int
}

type gormWriteSkewTest struct {
	db *gorm.DB
}

var _ crdb.WriteSkewTest = gormWriteSkewTest{}

func (t gormWriteSkewTest) Init(context.Context) error {
	t.db.AutoMigrate(&Account{})
	t.db.Create(&Account{ID: 1, Balance: 100})
	t.db.Create(&Account{ID: 2, Balance: 100})
	return t.db.Error
}

// ExecuteTx is part of the crdb.WriteSkewTest interface.
func (t gormWriteSkewTest) ExecuteTx(ctx context.Context, fn func(tx interface{}) error) error {
	return ExecuteTx(ctx, t.db, nil /* opts */, func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// GetBalances is part of the crdb.WriteSkewTest interface.
func (t gormWriteSkewTest) GetBalances(_ context.Context, txi interface{}) (int, int, error) {
	tx := txi.(*gorm.DB)

	var accounts []Account
	tx.Find(&accounts)
	if len(accounts) != 2 {
		return 0, 0, fmt.Errorf("expected two balances; got %d", len(accounts))
	}
	return accounts[0].Balance, accounts[1].Balance, nil
}

// UpdateBalance is part of the crdb.WriteSkewInterface.
func (t gormWriteSkewTest) UpdateBalance(
	_ context.Context, txi interface{}, accountID, delta int,
) error {
	tx := txi.(*gorm.DB)
	var acc Account
	tx.First(&acc, accountID)
	acc.Balance += delta
	return tx.Save(acc).Error
}
