package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetEnhancedMinidata represents the MSG_MHF_GET_ENHANCED_MINIDATA
type MsgMhfGetEnhancedMinidata struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetEnhancedMinidata) Opcode() network.PacketID {
	return network.MSG_MHF_GET_ENHANCED_MINIDATA
}

// Parse parses the packet from binary
func (m *MsgMhfGetEnhancedMinidata) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetEnhancedMinidata) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}