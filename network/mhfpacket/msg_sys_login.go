package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysLogin represents the MSG_SYS_LOGIN
type MsgSysLogin struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLogin) Opcode() network.PacketID {
	return network.MSG_SYS_LOGIN
}

// Parse parses the packet from binary
func (m *MsgSysLogin) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysLogin) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}