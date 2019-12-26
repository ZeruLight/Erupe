package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysLogout represents the MSG_SYS_LOGOUT
type MsgSysLogout struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLogout) Opcode() network.PacketID {
	return network.MSG_SYS_LOGOUT
}

// Parse parses the packet from binary
func (m *MsgSysLogout) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysLogout) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}