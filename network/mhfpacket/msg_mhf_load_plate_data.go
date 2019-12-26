package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadPlateData represents the MSG_MHF_LOAD_PLATE_DATA
type MsgMhfLoadPlateData struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadPlateData) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_PLATE_DATA
}

// Parse parses the packet from binary
func (m *MsgMhfLoadPlateData) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadPlateData) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}