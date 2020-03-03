package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve7C represents the MSG_SYS_reserve7C
type MsgSysReserve7C struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve7C) Opcode() network.PacketID {
	return network.MSG_SYS_reserve7C
}

// Parse parses the packet from binary
func (m *MsgSysReserve7C) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve7C) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
