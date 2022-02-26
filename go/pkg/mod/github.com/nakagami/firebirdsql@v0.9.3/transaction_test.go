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
	"database/sql"
	"testing"
	"time"
)

func TestTransaction(t *testing.T) {
	var n int
	test_dsn := GetTestDSN("test_transaction_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)
	if err != nil {
		t.Fatalf("Error sql.Open(): %v", err)
	}
	conn.Exec("CREATE TABLE test_trans (s varchar(2048))")
	conn.Close()

	time.Sleep(1 * time.Second)

	conn, err = sql.Open("firebirdsql", test_dsn)
	if err != nil {
		t.Fatalf("Error sql.Open(): %v", err)
	}
	err = conn.QueryRow("SELECT Count(*) FROM test_trans").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	if n != 0 {
		t.Fatalf("Incorrect count: %v", n)
	}
	conn.Exec("INSERT INTO test_trans (s) values ('A')")
	conn.Close()

	time.Sleep(1 * time.Second)

	conn, err = sql.Open("firebirdsql", test_dsn)
	if err != nil {
		t.Fatalf("sql.Open(): %v", err)
	}
	err = conn.QueryRow("SELECT Count(*) FROM test_trans").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	if n != 1 {
		t.Fatalf("Incorrect count: %v", n)
	}

	// Transaction
	tx, err := conn.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}

	// Rollback
	err = tx.QueryRow("SELECT Count(*) FROM test_trans").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	if n != 1 {
		t.Fatalf("Incorrect count: %v", n)
	}
	_, err = tx.Exec("INSERT INTO test_trans (s) values ('B')")
	if err != nil {
		t.Fatalf("Error Insert: %v", err)
	}
	err = tx.QueryRow("SELECT Count(*) FROM test_trans").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	if n != 2 {
		t.Fatalf("Incorrect count: %v", n)
	}
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("Error Rollback: %v", err)
	}

	tx, err = conn.Begin()
	err = tx.QueryRow("SELECT Count(*) FROM test_trans").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	if n != 1 {
		t.Fatalf("Incorrect count: %v", n)
	}

	// Commit
	_, err = tx.Exec("INSERT INTO test_trans (s) values ('C')")
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Error Commit: %v", err)
	}
	tx, err = conn.Begin()
	err = tx.QueryRow("SELECT Count(*) FROM test_trans").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	if n != 2 {
		t.Fatalf("Incorrect count: %v", n)
	}

	// without Commit (Need commit manually)
	_, err = tx.Exec("INSERT INTO test_trans (s) values ('D')")
	tx, err = conn.Begin()
	if err != nil {
		t.Fatalf("Error Begin: %v", err)
	}

	err = tx.QueryRow("SELECT Count(*) FROM test_trans").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	if n != 2 {
		t.Fatalf("Incorrect count: %v", n)
	}

	// Connection (autocommit)
	_, err = conn.Exec("INSERT INTO test_trans (s) values ('E')")
	if err != nil {
		t.Fatalf("Error Insert: %v", err)
	}
	conn.Close()

	time.Sleep(1 * time.Second)

	conn, err = sql.Open("firebirdsql", test_dsn)
	err = tx.QueryRow("SELECT Count(*) FROM test_trans").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	if n != 3 {
		t.Fatalf("Incorrect count: %v", n)
	}

	conn.Close()
}

func TestIssue35(t *testing.T) {
	test_dsn := GetTestDSN("test_issue35_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)

	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}

	tx, err := conn.Begin()

	if err != nil {
		t.Fatalf("Error Begin: %v", err)
	}

	err = tx.Commit()

	if err != nil {
		t.Fatalf("Error Commit: %v", err)
	}

	_, err = conn.Exec("CREATE TABLE test_issue35 (s varchar(2048))")

	if err != nil {
		t.Fatalf("Error CREATE TABLE: %v", err)
	}
	conn.Close()

	time.Sleep(1 * time.Second)

	conn, err = sql.Open("firebirdsql", test_dsn)
	var n int
	err = conn.QueryRow("SELECT Count(*) FROM test_issue35").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	if n != 0 {
		t.Fatalf("Incorrect count: %v", n)
	}

	conn.Close()
}

func TestIssue38(t *testing.T) {
	test_dsn := GetTestDSN("test_issue38_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)

	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}
	conn.Exec(`
        CREATE TABLE test_issue38 (
          id  INTEGER NOT NULL,
          key VARCHAR(64),
          value VARCHAR(64)
        )
    `)
	if err != nil {
		t.Fatalf("Error CREATE TABLE: %v", err)
	}
	conn.Close()

	time.Sleep(1 * time.Second)

	conn, err = sql.Open("firebirdsql", test_dsn)
	tx, err := conn.Begin()

	if err != nil {
		t.Fatalf("Error Begin: %v", err)
	}

	var rowId = sql.NullInt64{}

	err = tx.QueryRow(
		"INSERT INTO test_issue38 (id, key, value) VALUES (?, ?, ?) RETURNING id", 1, "testKey", "testValue").Scan(&rowId)
	if err == nil {
		t.Fatalf("'Dynamic SQL Error' is not occuerd.")
	}
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("Error Rollback: %v", err)
	}

	conn.Close()
}

func TestIssue39(t *testing.T) {
	conn, err := sql.Open("firebirdsql_createdb", GetTestDSN("test_issue39_"))
	tx, err := conn.Begin()

	if err != nil {
		t.Fatalf("Error Begin: %v", err)
	}
	var rowId = sql.NullInt64{}
	err = tx.QueryRow("select 5 / 0 from rdb$database").Scan(&rowId)
	if err == nil {
		t.Fatalf("'Dynamic SQL Error' is not occured.")
	}
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("broken transaction, but error is not occured.")
	}

	conn.Close()
}

func TestIssue67(t *testing.T) {
	test_dsn := GetTestDSN("test_issue67_")
	conn, _ := sql.Open("firebirdsql_createdb", test_dsn)
	var n int
	conn.QueryRow("SELECT Count(*) FROM rdb$relations").Scan(&n)
	err := conn.Close()
	if err != nil {
		t.Fatalf("Error Close: %v", err)
	}

	conn, _ = sql.Open("firebirdsql", test_dsn)
	tx, _ := conn.Begin()
	tx.QueryRow("SELECT Count(*) FROM rdb$relations").Scan(&n)

	tx.Commit()

	err = conn.Close()
	if err != nil {
		t.Fatalf("Error Close: %v", err)
	}

}

func TestIssue89(t *testing.T) {

	var noconn1, numberTrans, numberrelations int

	//	test transaction open on connection open
	test_dsn := GetTestDSN("test_issue89_")
	conn1, _ := sql.Open("firebirdsql_createdb", test_dsn)
	firebird_major_version := get_firebird_major_version(conn1)

	conn2, _ := sql.Open("firebirdsql", test_dsn)

	conn2.QueryRow("select count(*) from mon$transactions where mon$attachment_id <> current_connection").Scan(&numberTrans)

	if numberTrans > 0 {
		t.Fatalf("Transaction open without query runned")
	}

	conn2.Close()

	//	test if are more than 1 transaction open on first query
	conn1.QueryRow("select mon$attachment_id from mon$attachments where mon$attachment_id = current_connection").Scan(&noconn1)

	conn2, _ = sql.Open("firebirdsql", test_dsn)
	conn2.QueryRow("select count(*) from mon$transactions where mon$attachment_id <> current_connection").Scan(&numberTrans)

	if numberTrans > 1 {
		t.Fatalf("More than 1 transaction open")
	}

	conn1.Close()
	conn2.Close()

	//	test autocommit when rows is closed
	conn1, _ = sql.Open("firebirdsql", test_dsn)

	rows, _ := conn1.Query("select first 3 rdb$relation_id from rdb$relations")

	rows.Next()
	rows.Next()

	rows.Close()

	conn2, _ = sql.Open("firebirdsql", test_dsn)
	conn2.QueryRow("select count(*) from mon$transactions where mon$attachment_id <> current_connection").Scan(&numberTrans)

	if numberTrans > 0 {
		t.Fatalf("Autocommit don't work")
	}

	conn1.Close()
	conn2.Close()

	//	test autocommit on prepare statement
	conn1, _ = sql.Open("firebirdsql", test_dsn)
	stmt, _ := conn1.Prepare("select count(*) from rdb$relations")
	err := stmt.QueryRow().Scan(&numberrelations)

	if err != nil {
		t.Fatalf("Error QueryRow of Prepare: %v", err)
	}

	stmt.Close()

	stmt, _ = conn1.Prepare("select count(*) from rdb$relations")

	rows, _ = stmt.Query("select first 3 rdb$relation_id from rdb$relations")

	rows.Next()
	rows.Next()

	rows.Close()

	conn2, _ = sql.Open("firebirdsql", test_dsn)
	conn2.QueryRow("select count(*) from mon$transactions where mon$attachment_id <> current_connection").Scan(&numberTrans)

	if numberTrans > 0 {
		t.Fatalf("Autocommit in prepare don't work")
	}

	//	test autocommit on prepare statement
	conn1, _ = sql.Open("firebirdsql", test_dsn)
	conn1.Exec("create table testprepareinsert (id integer)")

	stmt, _ = conn1.Prepare("insert into testprepareinsert (id) values (?)")

	stmt.Exec(1)
	/*	_, err = stmt.Exec(2)
		if err == nil {
			t.Fatalf("Autocommit in prepare don't work")
		}
	*/
	stmt.Close()

	conn2, _ = sql.Open("firebirdsql", test_dsn)
	conn2.QueryRow("select count(*) from mon$transactions where mon$attachment_id <> current_connection").Scan(&numberTrans)

	if firebird_major_version < 3 && numberTrans > 0 {
		t.Fatalf("Autocommit in prepare don't work")
	}

	conn1, _ = sql.Open("firebirdsql", test_dsn)
	txp, _ := conn1.Begin()
	stmt, _ = txp.Prepare("insert into testprepareinsert (id) values (?)")

	for i := 1; i <= 6; i++ {
		_, err = stmt.Exec(i)
		if err != nil {
			t.Fatalf("Multiple execute of a prepared statement in same transaction don't work: %v", err)
		}
	}

	txp.Commit()
	conn2, _ = sql.Open("firebirdsql", test_dsn)
	conn2.QueryRow("select count(*) from mon$transactions where mon$attachment_id <> current_connection").Scan(&numberTrans)

	if firebird_major_version < 3 && numberTrans > 0 {
		t.Fatalf("Autocommit in prepare don't work")
	}

	// test transaction open after a commit of another transaction
	conn1, _ = sql.Open("firebirdsql", test_dsn)
	conn2, _ = sql.Open("firebirdsql", test_dsn)

	tx, err := conn1.Begin()

	if err != nil {
		t.Fatalf("Error opening new transaction: %v", err)
	}

	tx.QueryRow("select mon$attachment_id from mon$attachments where mon$attachment_id = current_connection").Scan(&noconn1)

	tx.Commit()

	err = conn1.QueryRow("select mon$attachment_id from mon$attachments where mon$attachment_id = current_connection").Scan(&noconn1)
	if err != nil {
		t.Fatalf("Error opening new transaction after last one committed or rollback: %v", err)
	}

	conn2, _ = sql.Open("firebirdsql", test_dsn)
	conn2.QueryRow("select count(*) from mon$transactions where mon$attachment_id <> current_connection").Scan(&numberTrans)

	if firebird_major_version < 3 && numberTrans > 0 {
		t.Fatalf("Transaction leaved open until close connection")
	}

	conn1.Close()
	conn2.Close()

}
