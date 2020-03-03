package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSetCaAchievementHist represents the MSG_MHF_SET_CA_ACHIEVEMENT_HIST
type MsgMhfSetCaAchievementHist struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetCaAchievementHist) Opcode() network.PacketID {
	return network.MSG_MHF_SET_CA_ACHIEVEMENT_HIST
}

// Parse parses the packet from binary
func (m *MsgMhfSetCaAchievementHist) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetCaAchievementHist) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
