package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetGemInfo represents the MSG_MHF_GET_GEM_INFO
type MsgMhfGetGemInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGemInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GEM_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetGemInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGemInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
