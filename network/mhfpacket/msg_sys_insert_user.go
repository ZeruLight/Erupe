package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysInsertUser represents the MSG_SYS_INSERT_USER
type MsgSysInsertUser struct {
	CharID uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysInsertUser) Opcode() network.PacketID {
	return network.MSG_SYS_INSERT_USER
}

// Parse parses the packet from binary
func (m *MsgSysInsertUser) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysInsertUser) Build(bf *byteframe.ByteFrame) error {
	bf.WriteUint32(m.CharID)
	return nil
}
