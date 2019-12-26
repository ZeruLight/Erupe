package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysSetUserBinary represents the MSG_SYS_SET_USER_BINARY
type MsgSysSetUserBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysSetUserBinary) Opcode() network.PacketID {
	return network.MSG_SYS_SET_USER_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysSetUserBinary) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysSetUserBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}