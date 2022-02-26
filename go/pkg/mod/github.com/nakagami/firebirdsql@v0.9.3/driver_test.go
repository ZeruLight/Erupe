/*******************************************************************************
The MIT License (MIT)

Copyright (c) 2013-2019 Hajime Nakagami

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
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func get_firebird_major_version(conn *sql.DB) int {
	var s string
	conn.QueryRow("SELECT rdb$get_context('SYSTEM', 'ENGINE_VERSION') from rdb$database").Scan(&s)
	major_version, _ := strconv.Atoi(s[:strings.Index(s, ".")])
	return major_version
}

func GetTestDSN(prefix string) string {
	var tmppath string
	randBytes := make([]byte, 16)
	rand.Read(randBytes)

	tmppath = filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+".fdb")
	if runtime.GOOS == "windows" {
		tmppath = "/" + tmppath
	}

	test_user := "sysdba"
	if isc_user := os.Getenv("ISC_USER"); isc_user != "" {
		test_user = isc_user
	}

	test_password := "masterkey"
	if isc_password := os.Getenv("ISC_PASSWORD"); isc_password != "" {
		test_password = isc_password
	}

	retorno := test_user + ":" + test_password + "@localhost:3050"
	return retorno + tmppath
}

func TestBasic(t *testing.T) {
	test_dsn := GetTestDSN("test_basic_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)

	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}
	var n int

	query := "SELECT Count(*) FROM rdb$relations where rdb$relation_name='FOO'"
	err = conn.QueryRow(query).Scan(&n)
	if err != nil {
		t.Fatalf("Error QueryRow: %v", err)
	}
	if n > 0 {
		conn.Exec("DROP TABLE foo")
	}

	query = `
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
    `
	conn.Exec(query)
	conn.Close()

	time.Sleep(1 * time.Second)

	conn, err = sql.Open("firebirdsql", test_dsn)
	_, err = conn.Exec("CREATE TABLE foo (a INTEGER)")
	if err == nil {
		t.Fatalf("Need metadata update error")
	}
	if !strings.Contains(err.Error(), "unsuccessful metadata update\n") {
		t.Fatalf("Bad message:%v", err.Error())
	}

	// 3 records insert
	conn.Exec("insert into foo(a, b, c, h) values (1, 'a', 'b','This is a memo')")
	conn.Exec("insert into foo(a, b, c, e, g, i, j) values (2, 'A', 'B', '1999-01-25', '00:00:01', 0.1, 0.1)")
	conn.Exec("insert into foo(a, b, c, e, g, i, j) values (3, 'X', 'Y', '2001-07-05', '00:01:02', 0.2, 0.2)")

	err = conn.QueryRow("select count(*) cnt from foo").Scan(&n)
	if err != nil {
		t.Fatalf("Error QueryRow: %v", err)
	}
	if n != 3 {
		t.Fatalf("Error bad record count: %v", n)
	}

	rows, err := conn.Query("select a, b, c, d, e, f, g, h, i, j from foo")
	columns, err := rows.Columns()
	if !reflect.DeepEqual(columns, []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}) {
		t.Fatalf("Columns() mismatch: %v", columns)
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
		// fmt.Println(a, b, c, d, e, f, g, h, i, j)
	}

	stmt, _ := conn.Prepare("select count(*) from foo where a=? and b=? and d=? and e=? and f=? and g=?")
	ep := time.Date(1967, 8, 11, 0, 0, 0, 0, time.Local)
	fp := time.Date(1967, 8, 11, 23, 45, 1, 0, time.Local)
	gp, err := time.Parse("15:04:05", "23:45:01")
	err = stmt.QueryRow(1, "a", -0.123, ep, fp, gp).Scan(&n)
	if err != nil {
		t.Fatalf("Error QueryRow: %v", err)
	}
	if n != 1 {
		t.Fatalf("Error bad record count: %v", n)
	}

	conn.Close()
}

func TestReturning(t *testing.T) {
	test_dsn := GetTestDSN("test_returning_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)
	if err != nil {
		t.Fatalf("Error sql.Open() : %v", err)
	}

	conn.Exec(`
        CREATE TABLE test_returning (
            f1 integer NOT NULL,
            f2 integer default 2,
            f3 varchar(20) default 'abc')`)

	conn.Close()

	time.Sleep(1 * time.Second)

	conn, err = sql.Open("firebirdsql", test_dsn)
	if err != nil {
		t.Fatalf("Error sql.Open() : %v", err)
	}

	for i := 0; i < 2; i++ {
		rows, err := conn.Query("INSERT INTO test_returning (f1) values (1) returning f2, f3")
		if err != nil {
			t.Fatalf("Error Insert returning : %v", err)
		}
		var f2 int
		var f3 string
		rows.Next()
		rows.Scan(&f2, &f3)
		if f2 != 2 || f3 != "abc" {
			t.Fatalf("Bad value insert returning: %v,%v", f2, f3)
		}
	}

	conn.Close()
}

func TestInsertBlobsWithParams(t *testing.T) {
	test_dsn := GetTestDSN("test_insert_blobs_with_params")
	conn, _ := sql.Open("firebirdsql_createdb", test_dsn)
	conn.Exec("CREATE TABLE test_blobs (f1 BLOB SUB_TYPE 0, f2 BLOB SUB_TYPE 1)")
	conn.Close()

	time.Sleep(1 * time.Second)

	conn, _ = sql.Open("firebirdsql", test_dsn)

	s0 := "Test Text"
	b0 := []byte{0, 1, 2, 3, 4, 13, 10, 5, 6, 7}
	if _, err := conn.Exec("INSERT INTO test_blobs (f1, f2) values (?, ?)", b0, s0); err != nil {
		t.Fatalf("Error inserting blobs with params: %v", err)
	}

	var s string
	var b []byte
	err := conn.QueryRow("SELECT f1, f2 from test_blobs").Scan(&b, &s)
	if err != nil {
		t.Fatalf("Error in query: %v", err)
	}
	if s != s0 {
		t.Fatalf("Text blob: expected <%s>, got <%s>", s0, s)
	}
	if !reflect.DeepEqual(b, b0) {
		t.Fatalf("Binary blob: expected <%v>, got <%v> (%s)", b0, b, string(b))
	}

	conn.Close()
}

func TestError(t *testing.T) {
	conn, err := sql.Open("firebirdsql_createdb", GetTestDSN("test_error_"))
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}
	_, err = conn.Exec("incorrect sql statement")
	if err == nil {
		t.Fatalf("Incorrect error")
	} else if err.Error() != "Dynamic SQL Error\nSQL error code = -104\nToken unknown - line 1, column 1\nincorrect\n" {
		t.Fatalf("Incorrect error: %v", err.Error())
	}
	conn.Close()
}

func TestRole(t *testing.T) {
	test_dsn := GetTestDSN("test_role_")
	conn1, err := sql.Open("firebirdsql_createdb", test_dsn)
	if err != nil {
		t.Fatalf("Error creating: %v", err)
	}
	conn1.Exec("CREATE TABLE test_role (f1 integer)")
	conn1.Exec("INSERT INTO test_role (f1) values (1)")
	_, err = conn1.Exec("CREATE ROLE DRIVERROLE")
	if err != nil {
		t.Fatalf("Error creating role: %v", err)
	}
	conn1.Exec("GRANT DRIVERROLE TO SYSDBA")
	if err != nil {
		t.Fatalf("Error creating role: %v", err)
	}
	conn1.Exec("GRANT SELECT ON test_role TO DRIVERROLE")
	if err != nil {
		t.Fatalf("Error granting right to role: %v", err)
	}
	conn1.Close()

	time.Sleep(1 * time.Second)

	conn2, err := sql.Open("firebirdsql", test_dsn+"?role=driverrole")
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}

	rows, err := conn2.Query("SELECT f1 FROM test_role")
	if err != nil {
		t.Fatalf("Error Query: %v", err)
	}

	for rows.Next() {
	}
	conn2.Close()
}

func TestInsertTimestamp(t *testing.T) {
	conn, err := sql.Open("firebirdsql_createdb", GetTestDSN("test_timestamp_"))
	if err != nil {
		t.Fatalf("Error creating: %v", err)
	}

	_, err = conn.Exec("CREATE TABLE TEST (VAL1 TIMESTAMP, VAL2 TIMESTAMP, VAL3 TIMESTAMP, VAL4 TIMESTAMP)")
	if err != nil {
		t.Fatalf("Error creating table: %v", err)
	}

	dt1 := time.Date(2015, 2, 9, 19, 25, 50, 740500000, time.Local)
	dt2 := "2015/2/9 19:25:50.7405"
	dt3 := "2015-2-9 19:25:50.7405"

	if _, err = conn.Exec("INSERT INTO TEST (VAL1, VAL2, VAL3, VAL4) VALUES (?, ?, ?, '2015/2/9 19:25:50.7405')", dt1, dt2, dt3); err != nil {
		t.Fatalf("Error executing insert: %s", err)
	}

	var rt1, rt2, rt3, rt4 time.Time

	err = conn.QueryRow("SELECT * FROM TEST").Scan(&rt1, &rt2, &rt3, &rt4)
	if err != nil {
		t.Fatalf("Unexpected error in select: %s", err)
	}

	if rt1 != dt1 {
		t.Errorf("Expected <%v>, got <%v>", dt1, rt1)
	}
	if rt2 != dt1 {
		t.Errorf("Expected <%v>, got <%v>", dt1, rt2)
	}
	if rt3 != dt1 {
		t.Errorf("Expected <%v>, got <%v>", dt1, rt3)
	}
	if rt4 != dt1 {
		t.Errorf("Expected <%v>, got <%v>", dt1, rt4)
	}
	conn.Close()
}

func TestBoolean(t *testing.T) {
	conn, err := sql.Open("firebirdsql_createdb", GetTestDSN("test_boolean_"))
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}

	firebird_major_version := get_firebird_major_version(conn)
	if firebird_major_version < 3 {
		return
	}

	var sql string
	var n int

	sql = "SELECT Count(*) FROM rdb$relations where rdb$relation_name='TEST_FB3'"
	err = conn.QueryRow(sql).Scan(&n)
	if err != nil {
		t.Fatalf("Error QueryRow: %v", err)
	}
	if n > 0 {
		conn.Exec("DROP TABLE test_fb3")
	}

	sql = `
        CREATE TABLE test_fb3 (
            b BOOLEAN
        )
    `
	conn.Exec(sql)
	conn.Exec("insert into test_fb3(b) values (true)")
	conn.Exec("insert into test_fb3(b) values (false)")
	var b bool
	err = conn.QueryRow("select * from test_fb3 where b is true").Scan(&b)
	if err != nil {
		t.Fatalf("Error QueryRow: %v", err)
	}
	if b != true {
		conn.Exec("Invalid boolean value")
	}
	err = conn.QueryRow("select * from test_fb3 where b is false").Scan(&b)
	if err != nil {
		t.Fatalf("Error QueryRow: %v", err)
	}
	if b != false {
		conn.Exec("Invalid boolean value")
	}

	stmt, _ := conn.Prepare("select * from test_fb3 where b=?")
	err = stmt.QueryRow(true).Scan(&b)
	if err != nil {
		t.Fatalf("Error QueryRow: %v", err)
	}
	if b != false {
		conn.Exec("Invalid boolean value")
	}

	conn.Close()
}

func TestDecFloat(t *testing.T) {
	conn, err := sql.Open("firebirdsql_createdb", GetTestDSN("test_decfloat_"))
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}

	firebird_major_version := get_firebird_major_version(conn)
	if firebird_major_version < 4 {
		return
	}

	query := `
        CREATE TABLE test_decfloat (
            i integer,
            d DECIMAL(20, 2),
            df16 DECFLOAT(16),
            df34 DECFLOAT(34),
            s varchar(32)
        )
    `
	conn.Exec(query)
	conn.Exec("insert into test_decfloat(i, d, df16, df34, s) values (1, 0.0, 0.0, 0.0, '0.0')")
	conn.Exec("insert into test_decfloat(i, d, df16, df34, s) values (2, 1.1, 1.1, 1.1, '1.1')")
	conn.Exec("insert into test_decfloat(i, d, df16, df34, s) values (3, 120.2, 120.2, 120.2, '120.2')")
	conn.Exec("insert into test_decfloat(i, d, df16, df34, s) values (4, -1.1, -1.1, -1.1, '-1.1')")
	conn.Exec("insert into test_decfloat(i, d, df16, df34, s) values (5, -120.2, -120.2, -120.2, '-120.2')")

	var n int
	err = conn.QueryRow("select count(*) cnt from test_decfloat").Scan(&n)
	if err != nil {
		t.Fatalf("Error QueryRow: %v", err)
	}
	if n != 5 {
		t.Fatalf("Error bad record count: %v", n)
	}

	rows, err := conn.Query("select df16, df34, s from test_decfloat order by i")

	var df16, df34 sql.NullFloat64
	var s string
	for rows.Next() {
		rows.Scan(&df16, &df34, &s)
		f, _ := strconv.ParseFloat(s, 64)
		df16v, _ := df16.Value()
		df34v, _ := df34.Value()

		if df16v != f || df34v != f {
			fmt.Printf("Error decfloat value : %v,%v,%v\n", df16v, df34v, f)
		}
	}

	conn.Close()
}

func TestTimeZone(t *testing.T) {
	conn, err := sql.Open("firebirdsql_createdb", GetTestDSN("test_timezone_")+"?timezone=Asia/Tokyo")
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}

	firebird_major_version := get_firebird_major_version(conn)
	if firebird_major_version < 4 {
		return
	}

	sql := `
            CREATE TABLE test_timezone (
                id INTEGER NOT NULL,
                a TIME WITH TIME ZONE DEFAULT '12:34:56',
                b TIMESTAMP WITH TIME ZONE DEFAULT '1967-08-11 23:45:01',
                PRIMARY KEY (id)
            )
    `
	conn.Exec(sql)
	conn.Exec("insert into test_timezone (id) values (0)")
	conn.Exec("insert into test_timezone (id, a, b) values (1, '12:34:56 Asia/Seoul', '1967-08-11 23:45:01.0000 Asia/Seoul')")
	conn.Exec("insert into test_timezone (id, a, b) values (2, '03:34:56 UTC', '1967-08-11 14:45:01.0000 UTC')")

	var id int
	var a time.Time
	var b time.Time
	rows, _ := conn.Query("select * from test_timezone")
	expected := []string{
		"0000-01-01 12:34:56 +0900 JST, 1967-08-11 23:45:01 +0900 JST",
		"0000-01-01 12:34:56 +0900 KST, 1967-08-11 23:45:01 +0900 KST",
		"0000-01-01 03:34:56 +0000 UTC, 1967-08-11 14:45:01 +0000 UTC"}

	for rows.Next() {
		rows.Scan(&id, &a, &b)
		s := fmt.Sprintf("%v, %v", a, b)
		if s != expected[id] {
			t.Fatalf("Incorrect result: %v", s)
		}
	}

	conn.Close()
}

func TestInt128(t *testing.T) {
	// https://github.com/nakagami/firebirdsql/issues/129
	conn, err := sql.Open("firebirdsql_createdb", GetTestDSN("test_int128_"))
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}

	firebird_major_version := get_firebird_major_version(conn)
	if firebird_major_version < 4 {
		return
	}

	sql := `
        CREATE TABLE test_int128 (
            i int128
        )
    `
	conn.Exec(sql)
	conn.Exec("insert into test_int128(i) values (170141183460469231731687303715884105727)")

	var i128 *big.Int
	err = conn.QueryRow("SELECT i FROM test_int128").Scan(&i128)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}

	var toCmp = new(big.Int)
	toCmp, _ = toCmp.SetString("170141183460469231731687303715884105727", 10)

	if i128.Cmp(toCmp) != 0 {
		t.Fatalf("INT128 Error: %v", i128)
	}

	conn.Close()
}

func TestNegativeInt128(t *testing.T) {
	conn, err := sql.Open("firebirdsql_createdb", GetTestDSN("test_negative_int128_"))
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}

	firebird_major_version := get_firebird_major_version(conn)
	if firebird_major_version < 4 {
		return
	}

	sql := `
        CREATE TABLE test_negative_int128 (
            i int128
        )
    `
	conn.Exec(sql)
	conn.Exec("insert into test_negative_int128(i) values (-170141183460469231731687303715884105727)")

	var i128 *big.Int
	err = conn.QueryRow("SELECT i FROM test_negative_int128").Scan(&i128)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}

	var toCmp = new(big.Int)
	toCmp, _ = toCmp.SetString("-170141183460469231731687303715884105727", 10)

	if i128.Cmp(toCmp) != 0 {
		t.Fatalf("Negative INT128 Error: %v", i128)
	}

	conn.Close()
}

func TestLegacyAuthWireCrypt(t *testing.T) {
	test_dsn := GetTestDSN("test_legacy_auth_")
	var n int
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}
	err = conn.Ping()
	if err != nil {
		t.Fatalf("Error ping: %v", err)
	}
	conn.Close()

	time.Sleep(1 * time.Second)

	conn, err = sql.Open("firebirdsql", test_dsn+"?auth_plugin_anme=Legacy_Auth")
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}
	err = conn.QueryRow("SELECT Count(*) FROM rdb$relations").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	conn.Close()

	conn, err = sql.Open("firebirdsql", test_dsn+"?wire_crypt=false")
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}
	err = conn.QueryRow("SELECT Count(*) FROM rdb$relations").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	conn.Close()

	conn, err = sql.Open("firebirdsql", test_dsn+"?auth_plugin_name=Legacy_Auth&wire_auth=true")
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}
	err = conn.QueryRow("SELECT Count(*) FROM rdb$relations").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	conn.Close()

	conn, err = sql.Open("firebirdsql", test_dsn+"?auth_plugin_name=Legacy_Auth&wire_auth=false")
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}
	err = conn.QueryRow("SELECT Count(*) FROM rdb$relations").Scan(&n)
	if err != nil {
		t.Fatalf("Error SELECT: %v", err)
	}
	conn.Close()
}

func TestErrorConnect(t *testing.T) {
	var n int
	conn, err := sql.Open("firebirdsql", "foo:bar@something_wrong_hostname:3050/dbname")
	if err != nil {
		t.Fatalf("Error occured at sql.Open()")
	}
	err = conn.QueryRow("SELECT Count(*) FROM rdb$relations").Scan(&n)
	if err == nil {
		t.Fatalf("Error not occured")
	}

	conn.Close()
}

func TestGoIssue44(t *testing.T) {
	conn, err := sql.Open("firebirdsql", "SomethingWrongConnectionString")
	err = conn.Ping()
	if err == nil {
		t.Fatalf("Error not occured")
	}
	conn.Close()
}

func TestGoIssue45(t *testing.T) {
	test_dsn := GetTestDSN("test_issue45_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)
	if err != nil {
		t.Fatalf("Error occured at sql.Open()")
	}

	conn.Exec(`
        CREATE TABLE person (
            name VARCHAR(60) NOT NULL,
            created TIMESTAMP
        )
    `)
	conn.Exec(`
        insert into person (name, created)
        values ('Giovanni', null)
    `)

	conn.Close()

	time.Sleep(1 * time.Second)

	conn, err = sql.Open("firebirdsql", test_dsn)

	// select null value
	type response struct {
		name    string
		created *time.Time
	}
	r := response{}

	err = conn.QueryRow(`
        select name, created from person
    `).Scan(&r.name, &r.created)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if r.created != nil {
		t.Fatalf("created is not nil")
	}

	// insert returning not null value
	err = conn.QueryRow(`
        insert into person (name, created)
        values ('Giovanni Gaspar', current_timestamp)
        returning name, created
    `).Scan(&r.name, &r.created)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if r.created == nil {
		t.Fatalf("created is nil")
	}

	// insert returning null value
	err = conn.QueryRow(`
        insert into person (name, created)
        values ('Nakagami', null)
        returning name, created
     `).Scan(
		&r.name, &r.created)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if r.created != nil {
		t.Fatalf("created is not nil")
	}

	conn.Close()
}

func TestGoIssue49(t *testing.T) {
	test_dsn := GetTestDSN("test_issue49_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)
	if err != nil {
		t.Fatalf("Error occured at sql.Open()")
	}
	defer conn.Close()

	sqlCreate := `
    CREATE TABLE NullTest (       
        name VARCHAR(60) NOT NULL,
        nullname VARCHAR(10),
        nullDate DATE,
        bug1 SMALLINT,
        bug2 INTEGER
    )
`
	conn.Exec(sqlCreate)

	//Worked
	sqlTest1 := `insert into NullTest (name, nullDate)values ('value', null)`
	_, err = conn.Exec(sqlTest1)
	if err != nil {
		t.Error(err)
	}
	//Worked
	sqlTest1 = `insert into NullTest (name, nullDate)values (?, ?)`
	_, err = conn.Exec(sqlTest1, "value", nil)
	if err != nil {
		t.Error(err)
	}

	//Failed
	sqlTest1 = `insert into NullTest (name, nullDate)values (?, ?)`
	_, err = conn.Exec(sqlTest1, "value", nil)
	if err != nil {
		t.Error(err)
	}

	// Failed
	sqlTest1 = `insert into NullTest (name, bug1) values ('value', ?)`
	_, err = conn.Exec(sqlTest1, nil)
	if err != nil {
		t.Error(err)
	}
	// Failed
	sqlTest1 = `insert into NullTest (name, bug1,bug2) values ('value', ?,?)`
	_, err = conn.Exec(sqlTest1, nil, nil)
	if err != nil {
		t.Error(err)
	}

	// must be failed!
	sqlTest1 = `insert into NullTest (name, bug1) values ('value', ?)`
	_, err = conn.Exec(sqlTest1)
	if err == nil {
		t.Error("Expected error!")
	}
}

func TestGoIssue53(t *testing.T) {
	timeout := time.Second * 40
	test_dsn := GetTestDSN("test_issue53_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)
	if err != nil {
		t.Fatalf("Error occured at sql.Open()")
	}
	defer conn.Close()

	tests := []int{
		31,
		BLOB_SEGMENT_SIZE,
		BLOB_SEGMENT_SIZE + 1,
		22*BLOB_SEGMENT_SIZE + 21,
		97*BLOB_SEGMENT_SIZE + 21,
	}

	conn.Exec(`CREATE TABLE BlobTest (bugField blob sub_type binary )`)

	for _, test := range tests {
		t.Run(fmt.Sprintf("%d", test), func(t *testing.T) {
			sqlDelete := `delete from BlobTest`
			_, err = conn.Exec(sqlDelete)
			if err != nil {
				t.Error(err)
			}

			sqlTest1 := `insert into BlobTest(bugField) values(?)`

			str := strings.Repeat("F", test)

			done := make(chan bool)
			go func(ch chan bool) {
				_, err = conn.Exec(sqlTest1, str)
				if err != nil {
					t.Error(err)
				}
				close(done)
			}(done)

			select {
			case <-done:
			case <-time.After(timeout):
				t.Fatal("Test timed out after ", timeout)
			}

			sqlget := `select bugField from BlobTest`
			rows, err := conn.Query(sqlget)
			if err != nil {
				t.Error(err)
				return
			}

			for rows.Next() {
				var buf []byte
				err = rows.Scan(&buf)
				if err != nil {
					t.Error(err)
				}
				if len(buf) != test {
					t.Errorf("Expected size blob %d, got %d", test, len(buf))
				}
			}

			rows.Close()

		})
	}
}
func TestGoIssue65(t *testing.T) {
	test_dsn := GetTestDSN("test_issue65_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)
	if err != nil {
		t.Fatalf("Error occured at sql.Open()")
	}
	defer conn.Close()

	conn.Exec(`CREATE TABLE FPI_MOVTO_MOVIMIENTOS
	(
	  RFCEMPRESA varchar(20) NOT NULL,
	  NOSUCURSAL integer NOT NULL,
	  TIPO integer NOT NULL,
	  SERIE varchar(5) NOT NULL,
	  NODOCTO integer NOT NULL,
	  LINEA integer NOT NULL,
	  CODART varchar(20),
	  NOMART varchar(80),
	  CLAVEPRODSERV varchar(10),
	  UNIDADCLAVE varchar(10),
	  UNIDADNOMBRE varchar(80),
	  CANT1 double precision,
	  CATN2 double precision,
	  PUNIT double precision,
	  MONTO double precision,
	  IMPTO1 double precision,
	  IMPTO2 double precision,
	  PIMPTO1 double precision,
	  PIMPTO2 double precision,
	  TIMPTO1 varchar(10),
	  TIMPTO2 varchar(10),
	  TFIMPTO1 varchar(10),
	  TFIMPTO2 varchar(10),
	  PDESCTO double precision,
	  IDESCTO double precision,
	  CONSTRAINT PXFPI_MOVTO_MOVIMIENTOS PRIMARY KEY (RFCEMPRESA,NOSUCURSAL,TIPO,SERIE,NODOCTO,LINEA)
	);`)

	//Worked
	sqlTest1 := `INSERT INTO FPI_MOVTO_MOVIMIENTOS (RFCEMPRESA, NOSUCURSAL, TIPO, SERIE, NODOCTO, LINEA, CODART, NOMART, CLAVEPRODSERV, UNIDADCLAVE, UNIDADNOMBRE, CANT1, CATN2, PUNIT, MONTO, IMPTO1, IMPTO2, PIMPTO1, PIMPTO2, TIMPTO1, TIMPTO2, TFIMPTO1, TFIMPTO2, PDESCTO, IDESCTO) VALUES ('p2', '0', '700', 'X', '1', '1', 'ART-001', 'PRUEBA DE ARTICULO', '01010101', 'ACT', 'Actividad', '10.000000', '0.000000', '2.500000', '25.000000', '4.000000', '0.000000', '16.000000', '0.000000', '002', '', 'Tasa', '', '0.000000', '0.000000');`
	_, err = conn.Exec(sqlTest1)
	if err != nil {
		t.Error(err)
	}

	sqlTest2 := "select doc.RFCEMPRESA, doc.NOSUCURSAL, doc.TIPO, doc.SERIE, doc.NODOCTO, doc.LINEA,\n" +
		"	doc.CODART, doc.NOMART, doc.CLAVEPRODSERV, doc.UNIDADCLAVE, doc.UNIDADNOMBRE, doc.CANT1,\n" +
		"	doc.CATN2, doc.PUNIT, doc.MONTO, doc.IMPTO1, doc.IMPTO2, doc.PIMPTO1, doc.PIMPTO2,\n" +
		"	doc.TIMPTO1, doc.TIMPTO2, doc.TFIMPTO1, doc.TFIMPTO2, doc.PDESCTO, doc.IDESCTO\n" +
		"from FPI_MOVTO_MOVIMIENTOS doc\n" +
		"where doc.RFCEMPRESA = 'p2' and doc.NOSUCURSAL = 0 and doc.TIPO = 700 and doc.SERIE = 'X' and doc.NODOCTO = 1 \n"
	movtos, err := conn.Query(sqlTest2)
	if err != nil {
		t.Error(err)
		return
	}

	existData := movtos.Next()
	if existData == false {
		t.Fatalf("Expecting Data")
	}
}

func TestGoIssue80(t *testing.T) {
	test_dsn := GetTestDSN("test_issue80_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)
	if err != nil {
		t.Fatalf("Error occured at sql.Open()")
	}
	defer conn.Close()

	query := `
        CREATE TABLE foo (
            a VARCHAR(10),
            b VARCHAR(10),
            c BIGINT,
            d INT,
            e INT,
            f INT,
            g INT,
            h INT,
            i INT,
            j INT,
            k INT,
            l INT,
            m INT,
            n INT
        )
    `

	_, err = conn.Exec(query)
	if err != nil {
		t.Error(err)
	}

	_, err = conn.Exec(
		"insert into foo(a, b, c, d, e, f, g, h, i, j, k, l, m, n) values (?, ?, ? ,?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		" ", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	if err != nil {
		t.Error(err)
	}

}

func TestIssue96(t *testing.T) {
	test_dsn := GetTestDSN("test_issue96_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}

	conn.Exec(`CREATE EXCEPTION EX_DATA_ERROR ''`)
	conn.Exec(`CREATE PROCEDURE EXCEPTION_PROC(in1 INTEGER)
        RETURNS (out1 INTEGER)
        AS
        BEGIN
          IF (IN1=1) THEN
          BEGIN
            EXCEPTION EX_DATA_ERROR 'data error';
          END
          out1 = in1;
          SUSPEND;
        END`)

	query := "SELECT * FROM exception_proc(1)"
	rows, err := conn.Query(query)
	if err != nil {
		t.Fatalf("Error Query: %v", err)
	}
	rows.Next()
	var n int
	err = rows.Scan(&n)
	if err == nil {
		t.Error("Error not occured")
	}

	conn.Close()
}

func TestGoIssue112(t *testing.T) {
	test_dsn := GetTestDSN("test_issue112_")
	conn, err := sql.Open("firebirdsql_createdb", test_dsn)
	if err != nil {
		t.Fatalf("Error occured at sql.Open()")
	}
	defer conn.Close()

	query := `
        CREATE TABLE foo (
            i BIGINT
        )
    `

	_, err = conn.Exec(query)
	if err != nil {
		t.Error(err)
	}

	var input_val, output_val int64
	input_val = 2147483648
	err = conn.QueryRow(`
        insert into foo (i)
        values (?)
        returning i
     `, input_val).Scan(
		&output_val)
	if err != nil {
		t.Error(err)
	}
	if input_val != output_val {
		t.Fatalf("%v != %v", input_val, output_val)
	}

}
