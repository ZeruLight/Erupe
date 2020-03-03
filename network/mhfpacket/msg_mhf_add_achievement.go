package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAddAchievement represents the MSG_MHF_ADD_ACHIEVEMENT
type MsgMhfAddAchievement struct {
	Unk0 uint8
	Unk1 uint16
	Unk2 uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddAchievement) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_ACHIEVEMENT
}

// Parse parses the packet from binary
func (m *MsgMhfAddAchievement) Parse(bf *byteframe.ByteFrame) error {
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint16()
	m.Unk2 = bf.ReadUint16()
	// doesn't expect a response
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddAchievement) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
