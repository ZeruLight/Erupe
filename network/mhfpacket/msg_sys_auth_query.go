package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysAuthQuery represents the MSG_SYS_AUTH_QUERY
type MsgSysAuthQuery struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysAuthQuery) Opcode() network.PacketID {
	return network.MSG_SYS_AUTH_QUERY
}

// Parse parses the packet from binary
func (m *MsgSysAuthQuery) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysAuthQuery) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
