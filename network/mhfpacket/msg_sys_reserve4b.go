package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve4B represents the MSG_SYS_reserve4B
type MsgSysReserve4B struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve4B) Opcode() network.PacketID {
	return network.MSG_SYS_reserve4B
}

// Parse parses the packet from binary
func (m *MsgSysReserve4B) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve4B) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
