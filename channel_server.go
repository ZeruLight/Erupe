package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

var loadDataCount int
var getPaperDataCount int

func handlePacket(cc *network.CryptConn, pkt []byte) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic.")
		}
	}()

	bf := byteframe.NewByteFrameFromBytes(pkt)
	opcode := network.PacketID(bf.ReadUint16())

	if opcode == network.MSG_SYS_EXTEND_THRESHOLD {
		opcode = network.PacketID(bf.ReadUint16())
	}

	switch opcode {
	case network.MSG_SYS_PING:
		ackHandle := bf.ReadUint32()
		_ = bf.ReadUint16()

		bfw := byteframe.NewByteFrame()
		bfw.WriteUint16(uint16(network.MSG_SYS_ACK))
		bfw.WriteUint32(ackHandle)
		bfw.WriteUint32(0)
		bfw.WriteUint32(0)
		cc.SendPacket(bfw.Data())
	case network.MSG_SYS_TIME:
		_ = bf.ReadUint8()
		timestamp := bf.ReadUint32() // unix timestamp, e.g. 1577105879

		bfw := byteframe.NewByteFrame()
		bfw.WriteUint16(uint16(network.MSG_SYS_TIME))
		bfw.WriteUint8(0)
		bfw.WriteUint32(timestamp)
		cc.SendPacket(bfw.Data())
	case network.MSG_SYS_LOGIN:
		ackHandle := bf.ReadUint32()
		charID0 := bf.ReadUint32()
		loginTokenNumber := bf.ReadUint32()
		hardcodedZero0 := bf.ReadUint16()
		requestVersion := bf.ReadUint16()
		charID1 := bf.ReadUint32()
		hardcodedZero1 := bf.ReadUint16()
		loginTokenLength := bf.ReadUint16() // hardcoded to 0x11
		loginTokenString := bf.ReadBytes(17)

		_ = ackHandle
		_ = charID0
		_ = loginTokenNumber
		_ = hardcodedZero0
		_ = requestVersion
		_ = charID1
		_ = hardcodedZero1
		_ = loginTokenLength
		_ = loginTokenString

		bfw := byteframe.NewByteFrame()
		bfw.WriteUint16(uint16(network.MSG_SYS_ACK))
		bfw.WriteUint32(ackHandle)
		bfw.WriteUint64(0x000000005E00B9C2) // Timestamp?
		cc.SendPacket(bfw.Data())

	case network.MSG_MHF_ENUMERATE_EVENT:
		fallthrough
	case network.MSG_MHF_ENUMERATE_QUEST:
		fallthrough
	case network.MSG_MHF_ENUMERATE_RANKING:
		fallthrough
	case network.MSG_MHF_READ_MERCENARY_W:
		fallthrough
	case network.MSG_MHF_GET_ETC_POINTS:
		fallthrough
	case network.MSG_MHF_READ_GUILDCARD:
		fallthrough
	case network.MSG_MHF_READ_BEAT_LEVEL:
		fallthrough
	case network.MSG_MHF_GET_EARTH_STATUS:
		fallthrough
	case network.MSG_MHF_GET_EARTH_VALUE:
		fallthrough
	case network.MSG_MHF_GET_WEEKLY_SCHEDULE:
		fallthrough
	case network.MSG_MHF_LIST_MEMBER:
		fallthrough
	case network.MSG_MHF_LOAD_PLATE_DATA:
		fallthrough
	case network.MSG_MHF_LOAD_PLATE_BOX:
		fallthrough
	case network.MSG_MHF_LOAD_FAVORITE_QUEST:
		fallthrough
	case network.MSG_MHF_LOAD_PARTNER:
		fallthrough
	case network.MSG_MHF_GET_TOWER_INFO:
		fallthrough
	case network.MSG_MHF_LOAD_OTOMO_AIROU:
		fallthrough
	case network.MSG_MHF_LOAD_DECO_MYSET:
		fallthrough
	case network.MSG_MHF_LOAD_HUNTER_NAVI:
		fallthrough
	case network.MSG_MHF_GET_UD_SCHEDULE:
		fallthrough
	case network.MSG_MHF_GET_UD_INFO:
		fallthrough
	case network.MSG_MHF_GET_UD_MONSTER_POINT:
		fallthrough
	case network.MSG_MHF_GET_RAND_FROM_TABLE:
		fallthrough
	case network.MSG_MHF_ACQUIRE_MONTHLY_REWARD:
		fallthrough
	case network.MSG_MHF_GET_RENGOKU_RANKING_RANK:
		fallthrough
	case network.MSG_MHF_LOAD_PLATE_MYSET:
		fallthrough
	case network.MSG_MHF_LOAD_RENGOKU_DATA:
		fallthrough
	case network.MSG_MHF_ENUMERATE_SHOP:
		fallthrough
	case network.MSG_MHF_LOAD_SCENARIO_DATA:
		fallthrough
	case network.MSG_MHF_GET_BOOST_TIME_LIMIT:
		fallthrough
	case network.MSG_MHF_GET_BOOST_RIGHT:
		fallthrough
	case network.MSG_MHF_GET_REWARD_SONG:
		fallthrough
	case network.MSG_MHF_GET_GACHA_POINT:
		fallthrough
	case network.MSG_MHF_GET_KOURYOU_POINT:
		fallthrough
	case network.MSG_MHF_GET_ENHANCED_MINIDATA:

		ackHandle := bf.ReadUint32()

		data, err := ioutil.ReadFile(fmt.Sprintf("bin_resp/%s_resp.bin", opcode.String()))
		if err != nil {
			panic(err)
		}

		bfw := byteframe.NewByteFrame()
		bfw.WriteUint16(uint16(network.MSG_SYS_ACK))
		bfw.WriteUint32(ackHandle)
		bfw.WriteBytes(data)
		cc.SendPacket(bfw.Data())

	case network.MSG_MHF_INFO_FESTA:
		ackHandle := bf.ReadUint32()
		_ = bf.ReadUint32()

		data, err := ioutil.ReadFile(fmt.Sprintf("bin_resp/%s_resp.bin", opcode.String()))
		if err != nil {
			panic(err)
		}

		bfw := byteframe.NewByteFrame()
		bfw.WriteUint16(uint16(network.MSG_SYS_ACK))
		bfw.WriteUint32(ackHandle)
		bfw.WriteBytes(data)
		cc.SendPacket(bfw.Data())

	case network.MSG_MHF_LOADDATA:
		ackHandle := bf.ReadUint32()

		data, err := ioutil.ReadFile(fmt.Sprintf("bin_resp/%s_resp%d.bin", opcode.String(), loadDataCount))
		if err != nil {
			panic(err)
		}

		bfw := byteframe.NewByteFrame()
		bfw.WriteUint16(uint16(network.MSG_SYS_ACK))
		bfw.WriteUint32(ackHandle)
		bfw.WriteBytes(data)
		cc.SendPacket(bfw.Data())

		loadDataCount++
		if loadDataCount > 1 {
			loadDataCount = 0
		}
	case network.MSG_MHF_GET_PAPER_DATA:
		ackHandle := bf.ReadUint32()

		data, err := ioutil.ReadFile(fmt.Sprintf("bin_resp/%s_resp%d.bin", opcode.String(), getPaperDataCount))
		if err != nil {
			panic(err)
		}

		bfw := byteframe.NewByteFrame()
		bfw.WriteUint16(uint16(network.MSG_SYS_ACK))
		bfw.WriteUint32(ackHandle)
		bfw.WriteBytes(data)
		cc.SendPacket(bfw.Data())

		getPaperDataCount++
		if getPaperDataCount > 7 {
			getPaperDataCount = 0
		}
	default:
		break
	}

	fmt.Printf("Opcode: %s\n", opcode)
	fmt.Printf("Data:\n%s\n", hex.Dump(pkt))

	remainingData := bf.DataFromCurrent()
	if len(remainingData) >= 2 && (opcode == network.MSG_SYS_TIME || opcode == network.MSG_MHF_INFO_FESTA) {
		handlePacket(cc, remainingData)
	}
}

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

		handlePacket(cc, pkt)

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
