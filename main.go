package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Andoryuuta/Erupe/signserver"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("Starting!")

	// Load the config.toml configuration.
	// TODO(Andoryuuta): implement config loading.

	// Create the postgres DB pool.
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=admin dbname=erupe sslmode=disable")
	if err != nil {
		panic(err)
	}

	// Test the DB connection.
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// Finally start our server(s).
	go serveLauncherHTML(":80")
	go doEntranceServer(":53310")

	signServer := signserver.NewServer(
		&signserver.Config{
			DB:         db,
			ListenAddr: ":53312",
		})
	go signServer.Listen()

	go doChannelServer(":54001")

	for {
		time.Sleep(1 * time.Second)
	}
}
