package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadGuildCooking represents the MSG_MHF_LOAD_GUILD_COOKING
type MsgMhfLoadGuildCooking struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadGuildCooking) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_GUILD_COOKING
}

// Parse parses the packet from binary
func (m *MsgMhfLoadGuildCooking) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadGuildCooking) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
