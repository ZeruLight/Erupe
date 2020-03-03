package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve20A represents the MSG_SYS_reserve20A
type MsgSysReserve20A struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve20A) Opcode() network.PacketID {
	return network.MSG_SYS_reserve20A
}

// Parse parses the packet from binary
func (m *MsgSysReserve20A) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve20A) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
