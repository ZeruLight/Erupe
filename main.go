package main

import (
	"fmt"
	"io"
	_ "time"

	"github.com/Andoryuuta/Erupe/network"
)

func main() {
	fmt.Println("Starting!")

	go serveLauncherHTML(":80")
	go doSignServer(":53312")

	for {
		time.Sleep(1 * time.Second)
	}
}
