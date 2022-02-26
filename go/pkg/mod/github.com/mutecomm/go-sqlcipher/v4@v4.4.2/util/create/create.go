package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func create(dbname, password string) error {
	dbnameWithDSN := dbname + fmt.Sprintf("?_pragma_key=%s&_pragma_cipher_page_size=4096", url.QueryEscape(password))
	db, err := sql.Open("sqlite3", dbnameWithDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE t1(a,b);")
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO t1(a,b) values('one for the money', 'two for the show');")
	return err
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s db_file password\n", os.Args[0])
	os.Exit(2)
}

func main() {
	if len(os.Args) != 3 {
		usage()
	}
	if err := create(os.Args[1], os.Args[2]); err != nil {
		fatal(err)
	}
}
