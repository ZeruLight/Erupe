package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAddGuildMissionCount represents the MSG_MHF_ADD_GUILD_MISSION_COUNT
type MsgMhfAddGuildMissionCount struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddGuildMissionCount) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_GUILD_MISSION_COUNT
}

// Parse parses the packet from binary
func (m *MsgMhfAddGuildMissionCount) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddGuildMissionCount) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
