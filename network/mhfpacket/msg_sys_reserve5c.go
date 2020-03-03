package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve5C represents the MSG_SYS_reserve5C
type MsgSysReserve5C struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve5C) Opcode() network.PacketID {
	return network.MSG_SYS_reserve5C
}

// Parse parses the packet from binary
func (m *MsgSysReserve5C) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve5C) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
