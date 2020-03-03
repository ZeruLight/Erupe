package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve20B represents the MSG_SYS_reserve20B
type MsgSysReserve20B struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve20B) Opcode() network.PacketID {
	return network.MSG_SYS_reserve20B
}

// Parse parses the packet from binary
func (m *MsgSysReserve20B) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve20B) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
