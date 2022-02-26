package sqlite3_test

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	sqlite3 "github.com/mutecomm/go-sqlcipher/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	db      *sql.DB
	testDir = "go-sqlcipher_test"
	tables  = `
CREATE TABLE KeyValueStore (
  KeyEntry   TEXT NOT NULL UNIQUE,
  ValueEntry TEXT NOT NULL
);`
)

func init() {
	// create DB
	key := url.QueryEscape("passphrase")
	tmpdir, err := ioutil.TempDir("", testDir)
	if err != nil {
		panic(err)
	}
	dbname := filepath.Join(tmpdir, "sqlcipher_test")
	dbnameWithDSN := dbname + fmt.Sprintf("?_pragma_key=%s&_pragma_cipher_page_size=4096", key)
	db, err = sql.Open("sqlite3", dbnameWithDSN)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(tables)
	if err != nil {
		panic(err)
	}
	db.Close()
	// make sure DB is encrypted
	encrypted, err := sqlite3.IsEncrypted(dbname)
	if err != nil {
		panic(err)
	}
	if !encrypted {
		panic(errors.New("go-sqlcipher: DB not encrypted"))
	}
	// open DB for testing
	db, err = sql.Open("sqlite3", dbnameWithDSN)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("SELECT count(*) FROM sqlite_master;")
	if err != nil {
		panic(err)
	}
}

var mapping = map[string]string{
	"foo": "one",
	"bar": "two",
	"baz": "three",
}

func TestSQLCipherParallelInsert(t *testing.T) {
	t.Parallel()
	insertValueQuery, err := db.Prepare("INSERT INTO KeyValueStore (KeyEntry, ValueEntry) VALUES (?, ?);")
	require.NoError(t, err)
	for key, value := range mapping {
		_, err := insertValueQuery.Exec(key, value)
		assert.NoError(t, err)
	}
}

func TestSQLCipherParallelSelect(t *testing.T) {
	t.Parallel()
	getValueQuery, err := db.Prepare("SELECT ValueEntry FROM KeyValueStore WHERE KeyEntry=?;")
	if err != nil {
		t.Fatal(err)
	}
	for key, value := range mapping {
		var val string
		err := getValueQuery.QueryRow(key).Scan(&val)
		if err != sql.ErrNoRows {
			if assert.NoError(t, err) {
				assert.Equal(t, value, val)
			}
		}
	}
}

func TestSQLCipherIsEncryptedFalse(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", testDir)
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	dbname := filepath.Join(tmpdir, "unencrypted.sqlite")
	db, err := sql.Open("sqlite3", dbname)
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec(tables)
	require.NoError(t, err)
	encrypted, err := sqlite3.IsEncrypted(dbname)
	if assert.NoError(t, err) {
		assert.False(t, encrypted)
	}
}

func TestSQLCipherIsEncryptedTrue(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", testDir)
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	dbname := filepath.Join(tmpdir, "encrypted.sqlite")
	var key [32]byte
	_, err = io.ReadFull(rand.Reader, key[:])
	require.NoError(t, err)
	dbnameWithDSN := dbname + fmt.Sprintf("?_pragma_key=x'%s'",
		hex.EncodeToString(key[:]))
	db, err := sql.Open("sqlite3", dbnameWithDSN)
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec(tables)
	require.NoError(t, err)
	encrypted, err := sqlite3.IsEncrypted(dbname)
	if assert.NoError(t, err) {
		assert.True(t, encrypted)
	}
}

func TestSQLCipher3DB(t *testing.T) {
	dbname := filepath.Join("testdata", "sqlcipher3.sqlite3")
	dbnameWithDSN := dbname + "?_pragma_key=passphrase&_pragma_cipher_page_size=4096"
	// make sure DB is encrypted
	encrypted, err := sqlite3.IsEncrypted(dbname)
	if err != nil {
		t.Fatal(err)
	}
	if !encrypted {
		t.Fatal("go-sqlcipher: DB not encrypted")
	}
	// open DB for testing
	db, err := sql.Open("sqlite3", dbnameWithDSN)
	if err != nil {
		t.Fatal(err)
	}
	// should fail
	_, err = db.Exec("SELECT count(*) FROM sqlite_master;")
	if err == nil {
		t.Fatal(errors.New("opening a SQLCipher 3 database with SQLCipher 4 should fail"))
	}
}

func TestSQLCipher4DB(t *testing.T) {
	dbname := filepath.Join("testdata", "sqlcipher4.sqlite3")
	dbnameWithDSN := dbname + "?_pragma_key=passphrase&_pragma_cipher_page_size=4096"
	// make sure DB is encrypted
	encrypted, err := sqlite3.IsEncrypted(dbname)
	if err != nil {
		t.Fatal(err)
	}
	if !encrypted {
		t.Fatal("go-sqlcipher: DB not encrypted")
	}
	// open DB for testing
	db, err := sql.Open("sqlite3", dbnameWithDSN)
	if err != nil {
		t.Fatal(err)
	}
	// should succeed
	_, err = db.Exec("SELECT count(*) FROM sqlite_master;")
	if err != nil {
		t.Fatal(err)
	}
}

func ExampleIsEncrypted() {
	// create random key
	var key [32]byte
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		log.Fatal(err)
	}
	// set DB name
	dbname := "go-sqlcipher.sqlite"
	dbnameWithDSN := dbname + fmt.Sprintf("?_pragma_key=x'%s'",
		hex.EncodeToString(key[:]))
	// create encrypted DB file
	db, err := sql.Open("sqlite3", dbnameWithDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(dbname)
	defer db.Close()
	// create table
	_, err = db.Exec("CREATE TABLE t(x INTEGER);")
	if err != nil {
		log.Fatal(err)
	}
	// make sure database is encrypted
	encrypted, err := sqlite3.IsEncrypted(dbname)
	if err != nil {
		log.Fatal(err)
	}
	if encrypted {
		fmt.Println("DB is encrypted")
	} else {
		fmt.Println("DB is unencrypted")
	}
	// Output:
	// DB is encrypted
}
