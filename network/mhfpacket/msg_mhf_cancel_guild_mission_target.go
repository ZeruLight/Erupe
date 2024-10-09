package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfCancelGuildMissionTarget represents the MSG_MHF_CANCEL_GUILD_MISSION_TARGET
type MsgMhfCancelGuildMissionTarget struct {
	AckHandle uint32
	MissionID uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCancelGuildMissionTarget) Opcode() network.PacketID {
	return network.MSG_MHF_CANCEL_GUILD_MISSION_TARGET
}

// Parse parses the packet from binary
func (m *MsgMhfCancelGuildMissionTarget) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.MissionID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCancelGuildMissionTarget) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
