package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve74 represents the MSG_SYS_reserve74
type MsgSysReserve74 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve74) Opcode() network.PacketID {
	return network.MSG_SYS_reserve74
}

// Parse parses the packet from binary
func (m *MsgSysReserve74) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve74) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
