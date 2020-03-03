package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve77 represents the MSG_SYS_reserve77
type MsgSysReserve77 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve77) Opcode() network.PacketID {
	return network.MSG_SYS_reserve77
}

// Parse parses the packet from binary
func (m *MsgSysReserve77) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve77) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
