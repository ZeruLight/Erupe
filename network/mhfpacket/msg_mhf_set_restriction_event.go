package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSetRestrictionEvent represents the MSG_MHF_SET_RESTRICTION_EVENT
type MsgMhfSetRestrictionEvent struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetRestrictionEvent) Opcode() network.PacketID {
	return network.MSG_MHF_SET_RESTRICTION_EVENT
}

// Parse parses the packet from binary
func (m *MsgMhfSetRestrictionEvent) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetRestrictionEvent) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
