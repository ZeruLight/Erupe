package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetCaUniqueId represents the MSG_MHF_GET_CA_UNIQUE_ID
type MsgMhfGetCaUniqueId struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetCaUniqueId) Opcode() network.PacketID {
	return network.MSG_MHF_GET_CA_UNIQUE_ID
}

// Parse parses the packet from binary
func (m *MsgMhfGetCaUniqueId) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetCaUniqueId) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}