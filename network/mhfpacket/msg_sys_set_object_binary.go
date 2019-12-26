package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysSetObjectBinary represents the MSG_SYS_SET_OBJECT_BINARY
type MsgSysSetObjectBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysSetObjectBinary) Opcode() network.PacketID {
	return network.MSG_SYS_SET_OBJECT_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysSetObjectBinary) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysSetObjectBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}