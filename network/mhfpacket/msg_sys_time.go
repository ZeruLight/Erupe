package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysTime represents the MSG_SYS_TIME
type MsgSysTime struct {
	Unk0      uint8
	Timestamp uint32 // unix timestamp, e.g. 1577105879
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysTime) Opcode() network.PacketID {
	return network.MSG_SYS_TIME
}

// Parse parses the packet from binary
func (m *MsgSysTime) Parse(bf *byteframe.ByteFrame) error {
	m.Unk0 = bf.ReadUint8()
	m.Timestamp = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysTime) Build(bf *byteframe.ByteFrame) error {
	bf.WriteUint8(m.Unk0)
	bf.WriteUint32(m.Timestamp)
	return nil
}
