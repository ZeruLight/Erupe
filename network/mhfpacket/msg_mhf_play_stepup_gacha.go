package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfPlayStepupGacha represents the MSG_MHF_PLAY_STEPUP_GACHA
type MsgMhfPlayStepupGacha struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPlayStepupGacha) Opcode() network.PacketID {
	return network.MSG_MHF_PLAY_STEPUP_GACHA
}

// Parse parses the packet from binary
func (m *MsgMhfPlayStepupGacha) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPlayStepupGacha) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
