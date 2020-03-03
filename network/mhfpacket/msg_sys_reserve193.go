package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve193 represents the MSG_SYS_reserve193
type MsgSysReserve193 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve193) Opcode() network.PacketID {
	return network.MSG_SYS_reserve193
}

// Parse parses the packet from binary
func (m *MsgSysReserve193) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve193) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
