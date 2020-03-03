package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSetRejectGuildScout represents the MSG_MHF_SET_REJECT_GUILD_SCOUT
type MsgMhfSetRejectGuildScout struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetRejectGuildScout) Opcode() network.PacketID {
	return network.MSG_MHF_SET_REJECT_GUILD_SCOUT
}

// Parse parses the packet from binary
func (m *MsgMhfSetRejectGuildScout) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetRejectGuildScout) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
