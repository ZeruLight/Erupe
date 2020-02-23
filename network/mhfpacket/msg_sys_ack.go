package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysAck represents the MSG_SYS_ACK
type MsgSysAck struct {
	AckHandle uint32
	AckData   []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysAck) Opcode() network.PacketID {
	return network.MSG_SYS_ACK
}

// Parse parses the packet from binary
func (m *MsgSysAck) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	panic("No way to parse without prior context as the packet doesn't include it's own length.")
}

// Build builds a binary packet from the current data.
func (m *MsgSysAck) Build(bf *byteframe.ByteFrame) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteBytes(m.AckData)
	return nil
}
