package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSavePlateData represents the MSG_MHF_SAVE_PLATE_DATA
type MsgMhfSavePlateData struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSavePlateData) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_PLATE_DATA
}

// Parse parses the packet from binary
func (m *MsgMhfSavePlateData) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSavePlateData) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}