package main

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

func handleChannelServerConnection(conn net.Conn) {
	fmt.Println("Channel server got connection!")
	// Unlike the sign and entrance server,
	// the client DOES NOT initalize the channel connection with 8 NULL bytes.

	cc := network.NewCryptConn(conn)
	for {
		pkt, err := cc.ReadPacket()
		if err != nil {
			return
		}

		bf := byteframe.NewByteFrameFromBytes(pkt)
		opcode := network.PacketID(bf.ReadUint16())
		fmt.Printf("Opcode: %s\n", opcode)
		fmt.Printf("Data:\n%s\n", hex.Dump(pkt))

	}
}

func doChannelServer(listenAddr string) {
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
		go handleChannelServerConnection(conn)
	}
}
