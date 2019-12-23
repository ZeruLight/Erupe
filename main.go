package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Starting!")

	go serveLauncherHTML(":80")
	go doEntranceServer(":53310")
	go doSignServer(":53312")
	go doChannelServer(":54001")

	for {
		time.Sleep(1 * time.Second)
	}
}
