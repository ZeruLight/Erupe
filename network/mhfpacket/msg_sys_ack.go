package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysAck represents the MSG_SYS_ACK
type MsgSysAck struct {
	AckHandle uint32
	Unk0      uint32
	Unk1      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysAck) Opcode() network.PacketID {
	return network.MSG_SYS_ACK
}

// Parse parses the packet from binary
func (m *MsgSysAck) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysAck) Build(bf *byteframe.ByteFrame) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint32(m.Unk0)
	bf.WriteUint32(m.Unk1)
	return nil
}
