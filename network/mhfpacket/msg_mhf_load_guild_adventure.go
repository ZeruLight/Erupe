package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadGuildAdventure represents the MSG_MHF_LOAD_GUILD_ADVENTURE
type MsgMhfLoadGuildAdventure struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadGuildAdventure) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_GUILD_ADVENTURE
}

// Parse parses the packet from binary
func (m *MsgMhfLoadGuildAdventure) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadGuildAdventure) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}