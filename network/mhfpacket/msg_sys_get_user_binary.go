package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysGetUserBinary represents the MSG_SYS_GET_USER_BINARY
type MsgSysGetUserBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysGetUserBinary) Opcode() network.PacketID {
	return network.MSG_SYS_GET_USER_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysGetUserBinary) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysGetUserBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}