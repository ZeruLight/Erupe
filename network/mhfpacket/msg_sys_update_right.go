package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysUpdateRight represents the MSG_SYS_UPDATE_RIGHT
type MsgSysUpdateRight struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysUpdateRight) Opcode() network.PacketID {
	return network.MSG_SYS_UPDATE_RIGHT
}

// Parse parses the packet from binary
func (m *MsgSysUpdateRight) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysUpdateRight) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}