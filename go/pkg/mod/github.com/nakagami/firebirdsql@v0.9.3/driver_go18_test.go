// +build go1.8

/*******************************************************************************
The MIT License (MIT)

Copyright (c) 2016-2019 Hajime Nakagami

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*******************************************************************************/

package firebirdsql

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestGo18(t *testing.T) {
	test_dsn := GetTestDSN("test_go18_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)

	if err != nil {
		t.Fatalf("Error sql.Open(): %v", err)
	}

	conn.Exec(`
        CREATE TABLE foo (
            a INTEGER NOT NULL,
            b VARCHAR(30) NOT NULL UNIQUE,
            c VARCHAR(1024),
            d DECIMAL(16,3) DEFAULT -0.123,
            e DATE DEFAULT '1967-08-11',
            f TIMESTAMP DEFAULT '1967-08-11 23:45:01',
            g TIME DEFAULT '23:45:01',
            h BLOB SUB_TYPE 1, 
            i DOUBLE PRECISION DEFAULT 0.0,
            j FLOAT DEFAULT 0.0,
            PRIMARY KEY (a),
            CONSTRAINT CHECK_A CHECK (a <> 0)
        )
    `)

	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}
	_, err = conn.Exec("insert into foo(a, b, c, h) values (1, 'a', 'b','This is a memo')")
	if err != nil {
		t.Fatalf("Error Insert1: %v", err)
	}
	_, err = conn.Exec("insert into foo(a, b, c, e, g, i, j) values (2, 'A', 'B', '1999-01-25', '00:00:01', 0.1, 0.1)")
	if err != nil {
		t.Fatalf("Error Insert2: %v", err)
	}

	conn.Close()

	time.Sleep(1 * time.Second)

	conn, err = sql.Open("firebirdsql", test_dsn)
	if err != nil {
		t.Fatalf("Error sql.Open(): %v", err)
	}

	ctx := context.Background()
	opts := &sql.TxOptions{sql.LevelDefault, true} // Default isolation leve and ReadOnly
	tx, err := conn.BeginTx(ctx, opts)
	if err != nil {
		t.Fatalf("Error BeginTx(): %v", err)
	}

	_, err = tx.Exec("insert into foo(a, b, c, e, g, i, j) values (3, 'X', 'Y', '2001-07-05', '00:01:02', 0.2, 0.2)")
	if err == nil {
		t.Fatalf("Error did not occured")
	} else if !strings.Contains(err.Error(), "read-only transaction") {
		t.Fatalf("Need read-only transaction error:%v", err)
	}

	var n int
	err = tx.QueryRow("select count(*) cnt from foo").Scan(&n)
	if err != nil {
		t.Fatalf("Error QueryRow: %v", err)
	}
	if n != 2 {
		t.Fatalf("Error bad record count: %v", n)
	}

	rows, err := tx.QueryContext(ctx, "select a, b, c, d, e, f, g, h, i, j from foo")
	ct, err := rows.ColumnTypes()
	var testColumnTypes = []struct {
		name     string
		typeName string
	}{
		{"A", "LONG"},
		{"B", "VARYING"},
		{"C", "VARYING"},
		{"D", "INT64"},
		{"E", "DATE"},
		{"F", "TIMESTAMP"},
		{"G", "TIME"},
		{"H", "BLOB"},
		{"I", "DOUBLE"},
		{"J", "FLOAT"},
	}

	for i, tct := range testColumnTypes {
		if ct[i].Name() != tct.name || ct[i].DatabaseTypeName() != tct.typeName {
			t.Fatalf("Error Column Type: %v", tct.name)
		}
	}

	var a int
	var b, c string
	var d float64
	var e time.Time
	var f time.Time
	var g time.Time
	var h []byte
	var i float64
	var j float32

	for rows.Next() {
		rows.Scan(&a, &b, &c, &d, &e, &f, &g, &h, &i, &j)
	}
	conn.Close()
}
