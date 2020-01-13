package channelserver

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"sync"

	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/Erupe/network/mhfpacket"
	"github.com/Andoryuuta/byteframe"
	"go.uber.org/zap"
)

// Session holds state for the channel server connection.
type Session struct {
	sync.Mutex
	logger    *zap.Logger
	server    *Server
	rawConn   net.Conn
	cryptConn *network.CryptConn
}

// NewSession creates a new Session type.
func NewSession(server *Server, conn net.Conn) *Session {
	s := &Session{
		logger:    server.logger,
		server:    server,
		rawConn:   conn,
		cryptConn: network.NewCryptConn(conn),
	}
	return s
}

// Start starts the session packet read&handle loop.
func (s *Session) Start() {
	go func() {
		s.logger.Info("Channel server got connection!")

		// Unlike the sign and entrance server,
		// the client DOES NOT initalize the channel connection with 8 NULL bytes.

		for {
			pkt, err := s.cryptConn.ReadPacket()
			if err != nil {
				s.logger.Warn("Error on channel server readpacket", zap.Error(err))
				return
			}

			s.handlePacketGroup(pkt)

		}
	}()
}

var loadDataCount int
var getPaperDataCount int

func (s *Session) handlePacketGroup(pktGroup []byte) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic.")
		}
	}()

	bf := byteframe.NewByteFrameFromBytes(pktGroup)
	opcode := network.PacketID(bf.ReadUint16())

	if opcode != network.MSG_SYS_END {
		fmt.Printf("Opcode: %s\n", opcode)
		fmt.Printf("Data:\n%s\n", hex.Dump(pktGroup))
	}

	switch opcode {
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
		s.cryptConn.SendPacket(bfw.Data())

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
		s.cryptConn.SendPacket(bfw.Data())

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
		s.cryptConn.SendPacket(bfw.Data())

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
		s.cryptConn.SendPacket(bfw.Data())

		getPaperDataCount++
		if getPaperDataCount > 7 {
			getPaperDataCount = 0
		}
	default:
		// Get the packet parser and handler for this opcode.
		mhfPkt := mhfpacket.FromOpcode(opcode)
		if mhfPkt == nil {
			fmt.Println("Got opcode which we don't know how to parse, can't parse anymore for this group")
			return
		}

		// Parse and handle the packet
		mhfPkt.Parse(bf)
		handlerTable[opcode](s, mhfPkt)
		break
	}

	// If there is more data on the stream that the .Parse method didn't read, then read another packet off it.
	remainingData := bf.DataFromCurrent()
	if len(remainingData) >= 2 && (opcode == network.MSG_SYS_TIME || opcode == network.MSG_MHF_INFO_FESTA || opcode == network.MSG_SYS_EXTEND_THRESHOLD) {
		s.handlePacketGroup(remainingData)
	}
}

/*
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

	fmt.Printf("Opcode: %s\n", opcode)
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
		fmt.Printf("Data:\n%s\n", hex.Dump(pkt))
		break
	}

	remainingData := bf.DataFromCurrent()
	if len(remainingData) >= 2 && (opcode == network.MSG_SYS_TIME || opcode == network.MSG_MHF_INFO_FESTA) {
		handlePacket(cc, remainingData)
	}
}
*/
