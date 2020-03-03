package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfDisplayedAchievement represents the MSG_MHF_DISPLAYED_ACHIEVEMENT
type MsgMhfDisplayedAchievement struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfDisplayedAchievement) Opcode() network.PacketID {
	return network.MSG_MHF_DISPLAYED_ACHIEVEMENT
}

// Parse parses the packet from binary
func (m *MsgMhfDisplayedAchievement) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfDisplayedAchievement) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
