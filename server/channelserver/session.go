package channelserver

import (
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/Andoryuuta/Erupe/common/stringstack"
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/Erupe/network/mhfpacket"
	"github.com/Andoryuuta/byteframe"
	"go.uber.org/zap"
)

// Session holds state for the channel server connection.
type Session struct {
	sync.Mutex
	logger      *zap.Logger
	server      *Server
	rawConn     net.Conn
	cryptConn   *network.CryptConn
	sendPackets chan []byte

	stageID          string
	stage            *Stage
	reservationStage *Stage // Required for the stateful MsgSysUnreserveStage packet.
	charID           uint32
	logKey           []byte

	// A stack containing the stage movement history (push on enter/move, pop on back)
	stageMoveStack *stringstack.StringStack
}

// NewSession creates a new Session type.
func NewSession(server *Server, conn net.Conn) *Session {
	s := &Session{
		logger:         server.logger.Named(conn.RemoteAddr().String()),
		server:         server,
		rawConn:        conn,
		cryptConn:      network.NewCryptConn(conn),
		sendPackets:    make(chan []byte, 20),
		stageMoveStack: stringstack.New(),
	}
	return s
}

// Start starts the session packet send and recv loop(s).
func (s *Session) Start() {
	go func() {
		s.logger.Info("Channel server got connection!", zap.String("remoteaddr", s.rawConn.RemoteAddr().String()))

		// Unlike the sign and entrance server,
		// the client DOES NOT initalize the channel connection with 8 NULL bytes.
		go s.sendLoop()
		s.recvLoop()
	}()
}

// QueueSend queues a packet (raw []byte) to be sent.
func (s *Session) QueueSend(data []byte) {
	if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.LogOutboundMessages {
		fmt.Printf("Sending To CharID: '%x'\n", s.charID)
		fmt.Printf("Sent Data:\n%s\n", hex.Dump(data))
	}

	s.sendPackets <- data
}

// QueueSendNonBlocking queues a packet (raw []byte) to be sent, dropping the packet entirely if the queue is full.
func (s *Session) QueueSendNonBlocking(data []byte) {
	select {
	case s.sendPackets <- data:
		// Enqueued properly.
	default:
		// Couldn't enqueue, likely something wrong with the connection.
		s.logger.Warn("Dropped packet for session because of full send buffer, something is probably wrong")
	}
}

// QueueSendMHF queues a MHFPacket to be sent.
func (s *Session) QueueSendMHF(pkt mhfpacket.MHFPacket) {
	// Make the header
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(pkt.Opcode()))

	// Build the packet onto the byteframe.
	pkt.Build(bf)

	// Queue it.
	s.QueueSend(bf.Data())
}

// QueueAck is a helper function to queue an MSG_SYS_ACK with the given ack handle and data.
func (s *Session) QueueAck(ackHandle uint32, data []byte) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(network.MSG_SYS_ACK))
	bf.WriteUint32(ackHandle)
	bf.WriteBytes(data)
	s.QueueSend(bf.Data())
}

func (s *Session) sendLoop() {
	for {
		// TODO(Andoryuuta): Test making this into a buffered channel and grouping the packet together before sending.
		rawPacket := <-s.sendPackets

		if rawPacket == nil {
			s.logger.Debug("Got nil from s.SendPackets, exiting send loop")
			return
		}

		// Make a copy of the data.
		terminatedPacket := make([]byte, len(rawPacket))
		copy(terminatedPacket, rawPacket)

		// Append the MSG_SYS_END tailing opcode.
		terminatedPacket = append(terminatedPacket, []byte{0x00, 0x10}...)

		s.cryptConn.SendPacket(terminatedPacket)
	}
}

func (s *Session) recvLoop() {
	for {
		pkt, err := s.cryptConn.ReadPacket()

		if err == io.EOF {
			s.logger.Info(fmt.Sprintf("Character(%d) disconnected", s.charID))
			logoutPlayer(s)
			return
		}

		if err != nil {
			s.logger.Warn("Error on ReadPacket, exiting recv loop", zap.Error(err))
			return
		}

		s.handlePacketGroup(pkt)
	}
}

func (s *Session) handlePacketGroup(pktGroup []byte) {
	// This shouldn't be needed, but it's better to recover and let the connection die than to panic the server.
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic ", r)
		}
	}()

	bf := byteframe.NewByteFrameFromBytes(pktGroup)
	opcode := network.PacketID(bf.ReadUint16())

	// Print any (non-common spam) packet opcodes and data.
	if opcode != network.MSG_SYS_END &&
		opcode != network.MSG_SYS_PING &&
		opcode != network.MSG_SYS_NOP &&
		opcode != network.MSG_SYS_TIME &&
		opcode != network.MSG_SYS_EXTEND_THRESHOLD {
		fmt.Printf("CharID: '%x'\n", s.charID)
		fmt.Printf("Opcode: %s\n", opcode)
		fmt.Printf("Data:\n%s\n", hex.Dump(pktGroup))
	}

	// Get the packet parser and handler for this opcode.
	mhfPkt := mhfpacket.FromOpcode(opcode)
	if mhfPkt == nil {
		fmt.Println("Got opcode which we don't know how to parse, can't parse anymore for this group")
		return
	}

	// Parse and handle the packet
	mhfPkt.Parse(bf)
	handlerTable[opcode](s, mhfPkt)

	// If there is more data on the stream that the .Parse method didn't read, then read another packet off it.
	remainingData := bf.DataFromCurrent()
	if len(remainingData) >= 2 {
		s.handlePacketGroup(remainingData)
	}
}
