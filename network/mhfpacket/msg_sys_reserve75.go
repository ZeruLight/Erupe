package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve75 represents the MSG_SYS_reserve75
type MsgSysReserve75 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve75) Opcode() network.PacketID {
	return network.MSG_SYS_reserve75
}

// Parse parses the packet from binary
func (m *MsgSysReserve75) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve75) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
