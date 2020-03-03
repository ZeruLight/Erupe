package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve205 represents the MSG_SYS_reserve205
type MsgSysReserve205 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve205) Opcode() network.PacketID {
	return network.MSG_SYS_reserve205
}

// Parse parses the packet from binary
func (m *MsgSysReserve205) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve205) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
