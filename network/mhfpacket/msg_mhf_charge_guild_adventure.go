package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfChargeGuildAdventure represents the MSG_MHF_CHARGE_GUILD_ADVENTURE
type MsgMhfChargeGuildAdventure struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfChargeGuildAdventure) Opcode() network.PacketID {
	return network.MSG_MHF_CHARGE_GUILD_ADVENTURE
}

// Parse parses the packet from binary
func (m *MsgMhfChargeGuildAdventure) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfChargeGuildAdventure) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
