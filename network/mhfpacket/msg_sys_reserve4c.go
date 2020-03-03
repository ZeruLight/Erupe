package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve4C represents the MSG_SYS_reserve4C
type MsgSysReserve4C struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve4C) Opcode() network.PacketID {
	return network.MSG_SYS_reserve4C
}

// Parse parses the packet from binary
func (m *MsgSysReserve4C) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve4C) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
