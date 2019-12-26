package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdGuildMapInfo represents the MSG_MHF_GET_UD_GUILD_MAP_INFO
type MsgMhfGetUdGuildMapInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdGuildMapInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_GUILD_MAP_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdGuildMapInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdGuildMapInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}