package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysTransBinary represents the MSG_SYS_TRANS_BINARY
type MsgSysTransBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysTransBinary) Opcode() network.PacketID {
	return network.MSG_SYS_TRANS_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysTransBinary) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysTransBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
