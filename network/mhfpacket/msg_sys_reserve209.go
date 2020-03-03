package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve209 represents the MSG_SYS_reserve209
type MsgSysReserve209 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve209) Opcode() network.PacketID {
	return network.MSG_SYS_reserve209
}

// Parse parses the packet from binary
func (m *MsgSysReserve209) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve209) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
