package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve207 represents the MSG_SYS_reserve207
type MsgSysReserve207 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve207) Opcode() network.PacketID {
	return network.MSG_SYS_reserve207
}

// Parse parses the packet from binary
func (m *MsgSysReserve207) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve207) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
