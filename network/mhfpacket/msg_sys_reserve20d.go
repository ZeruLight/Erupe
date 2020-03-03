package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve20D represents the MSG_SYS_reserve20D
type MsgSysReserve20D struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve20D) Opcode() network.PacketID {
	return network.MSG_SYS_reserve20D
}

// Parse parses the packet from binary
func (m *MsgSysReserve20D) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve20D) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
