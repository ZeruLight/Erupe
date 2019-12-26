package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSetEnhancedMinidata represents the MSG_MHF_SET_ENHANCED_MINIDATA
type MsgMhfSetEnhancedMinidata struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetEnhancedMinidata) Opcode() network.PacketID {
	return network.MSG_MHF_SET_ENHANCED_MINIDATA
}

// Parse parses the packet from binary
func (m *MsgMhfSetEnhancedMinidata) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetEnhancedMinidata) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}