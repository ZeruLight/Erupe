package main

import (
	"fmt"
	"io"
	"net"

	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

func paddedString(x string, size uint) []byte {
	out := make([]byte, size)
	copy(out, x)

	// Null terminate it.
	out[len(out)-1] = 0
	return out
}

func uint8PascalString(bf *byteframe.ByteFrame, x string) {
	bf.WriteUint8(uint8(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}

func uint16PascalString(bf *byteframe.ByteFrame, x string) {
	bf.WriteUint16(uint16(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}

func makeSignInResp(username string) []byte {
	bf := byteframe.NewByteFrame()

	bf.WriteUint8(1)                                   // resp_code
	bf.WriteUint8(0)                                   // file/patch server count
	bf.WriteUint8(1)                                   // entrance server count
	bf.WriteUint8(1)                                   // character count
	bf.WriteUint32(0xFFFFFFFF)                         // login_token_number
	bf.WriteBytes(paddedString("logintokenstrng", 16)) // login_token (16 byte padded string)
	bf.WriteUint32(1576761190)

	// file patch server PascalStrings here

	// Array(this.entrance_server_count, PascalString(Byte, "utf8")),
	uint8PascalString(bf, "localhost:53310")

	// Characters:
	bf.WriteUint32(1039336769) // character ID
	bf.WriteUint16(30)
	bf.WriteUint16(7)
	bf.WriteUint32(1576761172)
	bf.WriteUint8(0)
	bf.WriteUint8(0)
	bf.WriteUint8(0)
	bf.WriteUint8(0)
	bf.WriteBytes(paddedString("username", 16))
	bf.WriteBytes(paddedString("", 32))

	bf.WriteUint8(0)           // friends_list_count
	bf.WriteUint8(0)           // guild_members_count
	bf.WriteUint8(0)           // notice_count
	bf.WriteUint32(0xDEADBEEF) // some_last_played_character_id
	bf.WriteUint32(14)         // unk_flags
	uint8PascalString(bf, "")  // unk_data_blob PascalString

	bf.WriteUint16(51728)
	bf.WriteUint16(20000)
	uint16PascalString(bf, "1000672925")

	bf.WriteUint8(0)

	bf.WriteUint16(51729)
	bf.WriteUint16(1)
	bf.WriteUint16(20000)
	uint16PascalString(bf, "203.191.249.36:8080")

	bf.WriteUint32(1578905116)
	bf.WriteUint32(0)

	return bf.Data()
}

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
			return
		}

		bf := byteframe.NewByteFrameFromBytes(pkt)
		loginType := string(bf.ReadNullTerminatedBytes())
		username := string(bf.ReadNullTerminatedBytes())
		password := string(bf.ReadNullTerminatedBytes())
		unk := string(bf.ReadNullTerminatedBytes())
		fmt.Printf("Got signin, type: %s, username: %s, password %s, unk: %s", loginType, username, password, unk)

		resp := makeSignInResp(username)
		cc.SendPacket(resp)

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
