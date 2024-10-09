package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfGetAchievement represents the MSG_MHF_GET_ACHIEVEMENT
type MsgMhfGetAchievement struct {
	AckHandle uint32
	CharID    uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetAchievement) Opcode() network.PacketID {
	return network.MSG_MHF_GET_ACHIEVEMENT
}

// Parse parses the packet from binary
func (m *MsgMhfGetAchievement) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.CharID = bf.ReadUint32()
	bf.ReadUint32() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetAchievement) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
