package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysRightsReload represents the MSG_SYS_RIGHTS_RELOAD
type MsgSysRightsReload struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysRightsReload) Opcode() network.PacketID {
	return network.MSG_SYS_RIGHTS_RELOAD
}

// Parse parses the packet from binary
func (m *MsgSysRightsReload) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysRightsReload) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
