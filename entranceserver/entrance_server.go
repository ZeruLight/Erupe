package entranceserver

import (
	"encoding/hex"
	"fmt"
	"io"
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

		data := makeResp([]ServerInfo{
			ServerInfo{
				IP:                 net.ParseIP("127.0.0.1"),
				Unk2:               0,
				Type:               1,
				Season:             0,
				Unk6:               3,
				Name:               "AErupe Server in Go! @localhost",
				AllowedClientFlags: 4096,
				Channels: []ChannelInfo{
					ChannelInfo{
						Port:           54001,
						MaxPlayers:     100,
						CurrentPlayers: 3,
						Unk4:           0,
						Unk5:           0,
						Unk6:           0,
						Unk7:           0,
						Unk8:           0,
						Unk9:           0,
						Unk10:          319,
						Unk11:          248,
						Unk12:          159,
						Unk13:          12345,
					},
				},
			},
		})
		cc.SendPacket(data)

	}
}

func DoEntranceServer(listenAddr string) {
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
