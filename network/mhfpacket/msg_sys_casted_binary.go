package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCastedBinary represents the MSG_SYS_CASTED_BINARY
type MsgSysCastedBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCastedBinary) Opcode() network.PacketID {
	return network.MSG_SYS_CASTED_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysCastedBinary) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysCastedBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}