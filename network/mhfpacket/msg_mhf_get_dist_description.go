package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetDistDescription represents the MSG_MHF_GET_DIST_DESCRIPTION
type MsgMhfGetDistDescription struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetDistDescription) Opcode() network.PacketID {
	return network.MSG_MHF_GET_DIST_DESCRIPTION
}

// Parse parses the packet from binary
func (m *MsgMhfGetDistDescription) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetDistDescription) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
