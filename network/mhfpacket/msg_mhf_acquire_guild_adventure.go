package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcquireGuildAdventure represents the MSG_MHF_ACQUIRE_GUILD_ADVENTURE
type MsgMhfAcquireGuildAdventure struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireGuildAdventure) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_GUILD_ADVENTURE
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireGuildAdventure) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireGuildAdventure) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
