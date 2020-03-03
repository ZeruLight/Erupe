package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSetGuildManageRight represents the MSG_MHF_SET_GUILD_MANAGE_RIGHT
type MsgMhfSetGuildManageRight struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetGuildManageRight) Opcode() network.PacketID {
	return network.MSG_MHF_SET_GUILD_MANAGE_RIGHT
}

// Parse parses the packet from binary
func (m *MsgMhfSetGuildManageRight) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetGuildManageRight) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
