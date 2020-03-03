package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateInterior represents the MSG_MHF_UPDATE_INTERIOR
type MsgMhfUpdateInterior struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateInterior) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_INTERIOR
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateInterior) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateInterior) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
