package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetGuildMissionList represents the MSG_MHF_GET_GUILD_MISSION_LIST
type MsgMhfGetGuildMissionList struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGuildMissionList) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GUILD_MISSION_LIST
}

// Parse parses the packet from binary
func (m *MsgMhfGetGuildMissionList) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGuildMissionList) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
