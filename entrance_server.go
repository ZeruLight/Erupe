package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net"

	"github.com/Andoryuuta/Erupe/network"
)

func handleEntranceServerConnection(conn net.Conn) {
	// Client initalizes the connection with a one-time buffer of 8 NULL bytes.
	nullInit := make([]byte, 8)
	n, err := io.ReadFull(conn, nullInit)
	if err != nil {
		fmt.Println(err)
		return
	} else if n != len(nullInit) {
		fmt.Println("io.ReadFull couldn't read the full 8 byte init.")
		return
	}

	cc := network.NewCryptConn(conn)
	for {
		pkt, err := cc.ReadPacket()
		if err != nil {
			return
		}

		fmt.Printf("Got entrance server command:\n%s\n", hex.Dump(pkt))

		data, err := ioutil.ReadFile("custom_entrance_server_resp.bin")//("tw_server_list_resp.bin")
		if err != nil {
			print(err)
			return
		}
		cc.SendPacket(data)

	}
}

func doEntranceServer(listenAddr string) {
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go handleEntranceServerConnection(conn)
	}
}
