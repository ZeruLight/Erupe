package main

import (
	"fmt"
	"io"
	"net"

	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

func handleSignServerConnection(conn net.Conn) {
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
			panic(err)
		}

		bf := byteframe.NewByteFrameFromBytes(pkt)
		loginType := string(bf.ReadNullTerminatedBytes())
		username := string(bf.ReadNullTerminatedBytes())
		password := string(bf.ReadNullTerminatedBytes())
		unk := string(bf.ReadNullTerminatedBytes())
		fmt.Println("Got signin, type: %s, username: %s, password %s, unk: %s", loginType, username, password, unk)

		//fmt.Printf("Got:\n%s", hex.Dump(pkt))
	}
}

func doSignServer(listenAddr string) {
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
		go handleSignServerConnection(conn)
	}
}
