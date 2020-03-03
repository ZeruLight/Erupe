package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve4D represents the MSG_SYS_reserve4D
type MsgSysReserve4D struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve4D) Opcode() network.PacketID {
	return network.MSG_SYS_reserve4D
}

// Parse parses the packet from binary
func (m *MsgSysReserve4D) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve4D) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
