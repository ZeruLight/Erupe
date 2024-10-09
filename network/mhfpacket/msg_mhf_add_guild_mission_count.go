package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfAddGuildMissionCount represents the MSG_MHF_ADD_GUILD_MISSION_COUNT
type MsgMhfAddGuildMissionCount struct {
	AckHandle uint32
	MissionID uint32
	Count     uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddGuildMissionCount) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_GUILD_MISSION_COUNT
}

// Parse parses the packet from binary
func (m *MsgMhfAddGuildMissionCount) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.MissionID = bf.ReadUint32()
	m.Count = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddGuildMissionCount) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
