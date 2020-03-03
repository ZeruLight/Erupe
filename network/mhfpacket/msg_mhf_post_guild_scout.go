package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfPostGuildScout represents the MSG_MHF_POST_GUILD_SCOUT
type MsgMhfPostGuildScout struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostGuildScout) Opcode() network.PacketID {
	return network.MSG_MHF_POST_GUILD_SCOUT
}

// Parse parses the packet from binary
func (m *MsgMhfPostGuildScout) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostGuildScout) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
