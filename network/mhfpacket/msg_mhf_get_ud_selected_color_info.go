package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdSelectedColorInfo represents the MSG_MHF_GET_UD_SELECTED_COLOR_INFO
type MsgMhfGetUdSelectedColorInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdSelectedColorInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_SELECTED_COLOR_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdSelectedColorInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdSelectedColorInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}