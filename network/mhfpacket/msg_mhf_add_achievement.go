package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfAddAchievement represents the MSG_MHF_ADD_ACHIEVEMENT
type MsgMhfAddAchievement struct {
	AchievementID uint8
	Unk1          uint16
	Unk2          uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddAchievement) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_ACHIEVEMENT
}

// Parse parses the packet from binary
func (m *MsgMhfAddAchievement) Parse(bf *byteframe.ByteFrame) error {
	m.AchievementID = bf.ReadUint8()
	m.Unk1 = bf.ReadUint16()
	m.Unk2 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddAchievement) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
