package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSetKiju represents the MSG_MHF_SET_KIJU
type MsgMhfSetKiju struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetKiju) Opcode() network.PacketID {
	return network.MSG_MHF_SET_KIJU
}

// Parse parses the packet from binary
func (m *MsgMhfSetKiju) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetKiju) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
