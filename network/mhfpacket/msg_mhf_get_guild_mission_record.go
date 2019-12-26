package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetGuildMissionRecord represents the MSG_MHF_GET_GUILD_MISSION_RECORD
type MsgMhfGetGuildMissionRecord struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGuildMissionRecord) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GUILD_MISSION_RECORD
}

// Parse parses the packet from binary
func (m *MsgMhfGetGuildMissionRecord) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGuildMissionRecord) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}