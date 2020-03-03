package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve71 represents the MSG_SYS_reserve71
type MsgSysReserve71 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve71) Opcode() network.PacketID {
	return network.MSG_SYS_reserve71
}

// Parse parses the packet from binary
func (m *MsgSysReserve71) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve71) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
