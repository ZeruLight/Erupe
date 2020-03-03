package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateGuildIcon represents the MSG_MHF_UPDATE_GUILD_ICON
type MsgMhfUpdateGuildIcon struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateGuildIcon) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_GUILD_ICON
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateGuildIcon) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateGuildIcon) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
