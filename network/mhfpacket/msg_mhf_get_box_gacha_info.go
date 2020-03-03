package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetBoxGachaInfo represents the MSG_MHF_GET_BOX_GACHA_INFO
type MsgMhfGetBoxGachaInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetBoxGachaInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_BOX_GACHA_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetBoxGachaInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetBoxGachaInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
