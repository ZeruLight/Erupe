package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateGuacot represents the MSG_MHF_UPDATE_GUACOT
type MsgMhfUpdateGuacot struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateGuacot) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_GUACOT
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateGuacot) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateGuacot) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}