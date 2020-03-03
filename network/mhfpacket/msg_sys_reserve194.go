package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve194 represents the MSG_SYS_reserve194
type MsgSysReserve194 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve194) Opcode() network.PacketID {
	return network.MSG_SYS_reserve194
}

// Parse parses the packet from binary
func (m *MsgSysReserve194) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve194) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
