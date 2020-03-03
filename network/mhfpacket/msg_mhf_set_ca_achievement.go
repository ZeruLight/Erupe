package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSetCaAchievement represents the MSG_MHF_SET_CA_ACHIEVEMENT
type MsgMhfSetCaAchievement struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetCaAchievement) Opcode() network.PacketID {
	return network.MSG_MHF_SET_CA_ACHIEVEMENT
}

// Parse parses the packet from binary
func (m *MsgMhfSetCaAchievement) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetCaAchievement) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
