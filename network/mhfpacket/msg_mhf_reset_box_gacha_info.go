package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfResetBoxGachaInfo represents the MSG_MHF_RESET_BOX_GACHA_INFO
type MsgMhfResetBoxGachaInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfResetBoxGachaInfo) Opcode() network.PacketID {
	return network.MSG_MHF_RESET_BOX_GACHA_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfResetBoxGachaInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfResetBoxGachaInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
