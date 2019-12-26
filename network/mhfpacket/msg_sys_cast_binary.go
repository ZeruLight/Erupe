package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCastBinary represents the MSG_SYS_CAST_BINARY
type MsgSysCastBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCastBinary) Opcode() network.PacketID {
	return network.MSG_SYS_CAST_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysCastBinary) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysCastBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}