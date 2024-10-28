package channelserver

import (
	"encoding/binary"
	"encoding/hex"
	"erupe-ce/config"
	"erupe-ce/internal/constant"
	"erupe-ce/internal/system"
	"erupe-ce/network"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/database"
	"erupe-ce/utils/gametime"
	"erupe-ce/utils/logger"
	"erupe-ce/utils/mhfcourse"
	"erupe-ce/utils/stringstack"
	"fmt"
	"io"

	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Session holds state for the channel server connection.
type Session struct {
	sync.Mutex
	Logger      logger.Logger
	Server      *ChannelServer
	rawConn     net.Conn
	cryptConn   *network.CryptConn
	sendPackets chan mhfpacket.MHFPacket
	lastPacket  time.Time

	objectIndex      uint16
	userEnteredStage bool // If the user has entered a stage before
	stage            *system.Stage
	reservationStage *system.Stage // Required for the stateful MsgSysUnreserveStage packet.
	stagePass        string        // Temporary storage
	prevGuildID      uint32        // Stores the last GuildID used in InfoGuild
	CharID           uint32
	logKey           []byte
	sessionStart     int64
	courses          []mhfcourse.Course
	token            string
	kqf              []byte
	kqfOverride      bool

	semaphore     *Semaphore // Required for the stateful MsgSysUnreserveStage packet.
	semaphoreMode bool
	semaphoreID   []uint16

	// A stack containing the stage movement history (push on enter/move, pop on back)
	stageMoveStack *stringstack.StringStack

	// Accumulated index used for identifying mail for a client
	// I'm not certain why this is used, but since the client is sending it
	// I want to rely on it for now as it might be important later.
	mailAccIndex uint8
	// Contains the mail list that maps accumulated indexes to mail IDs
	mailList []int

	// For Debuging
	Name     string
	closed   bool
	ackStart map[uint32]time.Time
}

// NewSession creates a new Session type.
func NewSession(server *ChannelServer, conn net.Conn) *Session {
	s := &Session{
		Logger:         server.logger.Named(conn.RemoteAddr().String()),
		Server:         server,
		rawConn:        conn,
		cryptConn:      network.NewCryptConn(conn),
		sendPackets:    make(chan mhfpacket.MHFPacket, 20),
		lastPacket:     time.Now(),
		sessionStart:   gametime.TimeAdjusted().Unix(),
		stageMoveStack: stringstack.New(),
		ackStart:       make(map[uint32]time.Time),
		semaphoreID:    make([]uint16, 2),
	}
	s.SetObjectID()
	return s
}

// Start starts the session packet send and recv loop(s).
func (s *Session) Start() {
	s.Logger.Debug("New connection", zap.String("RemoteAddr", s.rawConn.RemoteAddr().String()))
	// Unlike the sign and entrance server,
	// the client DOES NOT initalize the channel connection with 8 NULL bytes.
	go s.sendLoop()
	go s.recvLoop()
}

// QueueSendMHF queues a MHFPacket to be sent.
func (s *Session) QueueSendMHF(pkt mhfpacket.MHFPacket) {
	select {
	case s.sendPackets <- pkt:
	default:
		s.Logger.Warn("Packet queue too full, dropping!")
	}
}

func (s *Session) sendLoop() {
	var pkt mhfpacket.MHFPacket
	var buffer []byte
	end := &mhfpacket.MsgSysEnd{}
	for {
		if s.closed {
			return
		}
		for len(s.sendPackets) > 0 {
			pkt = <-s.sendPackets
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(uint16(pkt.Opcode()))
			pkt.Build(bf)
			s.logMessage(uint16(pkt.Opcode()), bf.Data()[2:], "Server", s.Name)
			buffer = append(buffer, bf.Data()...)
		}
		bf := byteframe.NewByteFrame()
		bf.WriteUint16(uint16(end.Opcode()))
		buffer = append(buffer, bf.Data()...)
		if len(buffer) > 0 {
			err := s.cryptConn.SendPacket(buffer)
			if err != nil {
				s.Logger.Warn("Failed to send packet")
			}
			buffer = buffer[:0]
		}
		time.Sleep(time.Duration(config.GetConfig().LoopDelay) * time.Millisecond)
	}
}

func (s *Session) recvLoop() {
	for {
		if time.Now().Add(-30 * time.Second).After(s.lastPacket) {
			logoutPlayer(s)
			return
		}
		if s.closed {
			logoutPlayer(s)
			return
		}
		pkt, err := s.cryptConn.ReadPacket()
		if err == io.EOF {
			s.Logger.Info(fmt.Sprintf("[%s] Disconnected", s.Name))
			logoutPlayer(s)
			return
		} else if err != nil {
			s.Logger.Warn("Error on ReadPacket, exiting recv loop", zap.Error(err))
			logoutPlayer(s)
			return
		}
		s.handlePacketGroup(pkt)
		time.Sleep(time.Duration(config.GetConfig().LoopDelay) * time.Millisecond)
	}
}

func (s *Session) handlePacketGroup(pktGroup []byte) {
	s.lastPacket = time.Now()
	bf := byteframe.NewByteFrameFromBytes(pktGroup)
	opcodeUint16 := bf.ReadUint16()
	if len(bf.Data()) >= 6 {
		s.ackStart[bf.ReadUint32()] = time.Now()
		bf.Seek(2, io.SeekStart)
	}
	opcode := network.PacketID(opcodeUint16)

	// This shouldn't be needed, but it's better to recover and let the connection die than to panic the server.
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[%s]", s.Name)
			fmt.Println("Recovered from panic", r)
		}
	}()

	s.logMessage(opcodeUint16, pktGroup, s.Name, "Server")

	if opcode == network.MSG_SYS_LOGOUT {
		s.closed = true
		return
	}
	// Get the packet parser and handler for this opcode.
	mhfPkt := mhfpacket.FromOpcode(opcode)
	if mhfPkt == nil {
		fmt.Println("Got opcode which we don't know how to parse, can't parse anymore for this group")
		return
	}
	// Parse the packet.
	err := mhfPkt.Parse(bf)
	if err != nil {
		fmt.Printf("\n!!! [%s] %s NOT IMPLEMENTED !!! \n\n\n", s.Name, opcode)
		return
	}
	database, err := database.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	// Handle the packet.
	handlerTable[opcode](s, database, mhfPkt)
	// If there is more data on the stream that the .Parse method didn't read, then read another packet off it.
	remainingData := bf.DataFromCurrent()
	if len(remainingData) >= 2 {
		s.handlePacketGroup(remainingData)
	}
}

func ignored(opcode network.PacketID) bool {
	ignoreList := []network.PacketID{
		network.MSG_SYS_END,
		network.MSG_SYS_PING,
		network.MSG_SYS_NOP,
		network.MSG_SYS_TIME,
		network.MSG_SYS_EXTEND_THRESHOLD,
		network.MSG_SYS_POSITION_OBJECT,
		network.MSG_MHF_SAVEDATA,
	}
	set := make(map[network.PacketID]struct{}, len(ignoreList))
	for _, s := range ignoreList {
		set[s] = struct{}{}
	}
	_, r := set[opcode]
	return r
}

func (s *Session) logMessage(opcode uint16, data []byte, sender string, recipient string) {
	if sender == "Server" && !config.GetConfig().DebugOptions.LogOutboundMessages {
		return
	} else if sender != "Server" && !config.GetConfig().DebugOptions.LogInboundMessages {
		return
	}

	opcodePID := network.PacketID(opcode)
	if ignored(opcodePID) {
		return
	}
	var ackHandle uint32
	if len(data) >= 6 {
		ackHandle = binary.BigEndian.Uint32(data[2:6])
	}
	if t, ok := s.ackStart[ackHandle]; ok {
		fmt.Printf("[%s] -> [%s] (%fs)\n", sender, recipient, float64(time.Now().UnixNano()-t.UnixNano())/1000000000)
	} else {
		fmt.Printf("[%s] -> [%s]\n", sender, recipient)
	}
	fmt.Printf("Opcode: (Dec: %d Hex: 0x%04X Name: %s) \n", opcode, opcode, opcodePID)
	if config.GetConfig().DebugOptions.LogMessageData {
		if len(data) <= config.GetConfig().DebugOptions.MaxHexdumpLength {
			fmt.Printf("Data [%d bytes]:\n%s\n", len(data), hex.Dump(data))
		} else {
			fmt.Printf("Data [%d bytes]: (Too long!)\n\n", len(data))
		}
	} else {
		fmt.Printf("\n")
	}
}

func (s *Session) sendMessage(message string) {
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	msgBinChat := &binpacket.MsgBinChat{
		Unk0:       0,
		Type:       5,
		Flags:      0x80,
		Message:    message,
		SenderName: "Erupe",
	}
	msgBinChat.Build(bf)
	castedBin := &mhfpacket.MsgSysCastedBinary{
		CharID:         0,
		MessageType:    constant.BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}
	s.QueueSendMHF(castedBin)
}

func (s *Session) SetObjectID() {
	for i := uint16(1); i < 127; i++ {
		exists := false
		for _, j := range s.Server.objectIDs {
			if i == j {
				exists = true
				break
			}
		}
		if !exists {
			s.Server.objectIDs[s] = i
			return
		}
	}
	s.Server.objectIDs[s] = 0
}

func (s *Session) NextObjectID() uint32 {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(s.Server.objectIDs[s])
	s.objectIndex++
	bf.WriteUint16(s.objectIndex)
	bf.Seek(0, 0)
	return bf.ReadUint32()
}

func (s *Session) GetSemaphoreID() uint32 {
	if s.semaphoreMode {
		return 0x000E0000 + uint32(s.semaphoreID[1])
	} else {
		return 0x000F0000 + uint32(s.semaphoreID[0])
	}
}

func (s *Session) isOp() bool {
	var op bool
	db, err := database.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	err = db.QueryRow(`SELECT op FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`, s.CharID).Scan(&op)
	if err == nil && op {
		return true
	}
	return false
}

func (s *Session) DoAckEarthSucceed(ackHandle uint32, data []*byteframe.ByteFrame) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(config.GetConfig().EarthID))
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(uint32(len(data)))
	for i := range data {
		bf.WriteBytes(data[i].Data())
	}
	s.DoAckBufSucceed(ackHandle, bf.Data())
}

func (s *Session) DoAckBufSucceed(ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: true,
		ErrorCode:        0,
		AckData:          data,
	})
}

func (s *Session) DoAckBufFail(ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: true,
		ErrorCode:        1,
		AckData:          data,
	})
}

func (s *Session) DoAckSimpleSucceed(ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: false,
		ErrorCode:        0,
		AckData:          data,
	})
}

func (s *Session) DoAckSimpleFail(ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: false,
		ErrorCode:        1,
		AckData:          data,
	})
}
func (s *Session) GetCharID() uint32 {
	return s.CharID
}
func (s *Session) GetName() string {
	return s.Name
}

func (s *Session) Getkqf() []byte {
	return s.kqf
}

func (s *Session) GetkqfOverride() bool {
	return s.kqfOverride
}

func (s *Session) Setkqf(kqf []byte) {
	s.kqf = kqf
}
