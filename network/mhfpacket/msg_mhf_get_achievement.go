package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetAchievement represents the MSG_MHF_GET_ACHIEVEMENT
type MsgMhfGetAchievement struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetAchievement) Opcode() network.PacketID {
	return network.MSG_MHF_GET_ACHIEVEMENT
}

// Parse parses the packet from binary
func (m *MsgMhfGetAchievement) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetAchievement) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
