package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func selectFromDB(dbname, password string) error {
	dbnameWithDSN := dbname + fmt.Sprintf("?_pragma_key=%s&_pragma_cipher_page_size=4096", url.QueryEscape(password))
	db, err := sql.Open("sqlite3", dbnameWithDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	var a, b string
	row := db.QueryRow("SELECT * FROM t1;")
	err = row.Scan(&a, &b)
	if err != nil {
		return err
	}
	fmt.Printf("%s, %s\n", a, b)
	return nil
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
	if err := selectFromDB(os.Args[1], os.Args[2]); err != nil {
		fatal(err)
	}
}
