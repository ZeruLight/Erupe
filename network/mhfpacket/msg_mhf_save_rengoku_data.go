package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSaveRengokuData represents the MSG_MHF_SAVE_RENGOKU_DATA
type MsgMhfSaveRengokuData struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSaveRengokuData) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_RENGOKU_DATA
}

// Parse parses the packet from binary
func (m *MsgMhfSaveRengokuData) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSaveRengokuData) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}