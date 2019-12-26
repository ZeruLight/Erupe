package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetPaperData represents the MSG_MHF_GET_PAPER_DATA
type MsgMhfGetPaperData struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetPaperData) Opcode() network.PacketID {
	return network.MSG_MHF_GET_PAPER_DATA
}

// Parse parses the packet from binary
func (m *MsgMhfGetPaperData) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetPaperData) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}