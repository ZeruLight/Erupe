package binpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

type ChatTargetType uint16

const (
	CHAT_TARGET_PRIVATE = 0x05
	CHAT_TARGET_PARTY   = 0x04
)

type MsgBinTargetedChatMessage struct {
	// I can't see a reason if this is indeed the number of targets, that
	// it should use 2 bytes
	TargetCount    uint16
	TargetCharIDs  []uint32
	TargetType     uint16
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgBinTargetedChatMessage) Opcode() network.PacketID {
	return network.MSG_SYS_CAST_BINARY
}

func (m *MsgBinTargetedChatMessage) Parse(bf *byteframe.ByteFrame) error {
	m.TargetCount = bf.ReadUint16()
	i := uint16(0)

	m.TargetCharIDs = make([]uint32, m.TargetCount)

	for ; i < m.TargetCount; i++ {
		m.TargetCharIDs[i] = bf.ReadUint32()
	}

	m.TargetType = bf.ReadUint16()
	m.RawDataPayload = bf.DataFromCurrent()

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgBinTargetedChatMessage) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
