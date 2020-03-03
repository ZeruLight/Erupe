package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateForceGuildRank represents the MSG_MHF_UPDATE_FORCE_GUILD_RANK
type MsgMhfUpdateForceGuildRank struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateForceGuildRank) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_FORCE_GUILD_RANK
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateForceGuildRank) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateForceGuildRank) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
