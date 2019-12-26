package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSaveMezfesData represents the MSG_MHF_SAVE_MEZFES_DATA
type MsgMhfSaveMezfesData struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSaveMezfesData) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_MEZFES_DATA
}

// Parse parses the packet from binary
func (m *MsgMhfSaveMezfesData) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSaveMezfesData) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}