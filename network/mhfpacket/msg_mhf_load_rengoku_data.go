package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadRengokuData represents the MSG_MHF_LOAD_RENGOKU_DATA
type MsgMhfLoadRengokuData struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadRengokuData) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_RENGOKU_DATA
}

// Parse parses the packet from binary
func (m *MsgMhfLoadRengokuData) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadRengokuData) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}