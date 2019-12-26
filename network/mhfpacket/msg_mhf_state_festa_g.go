package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfStateFestaG represents the MSG_MHF_STATE_FESTA_G
type MsgMhfStateFestaG struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfStateFestaG) Opcode() network.PacketID {
	return network.MSG_MHF_STATE_FESTA_G
}

// Parse parses the packet from binary
func (m *MsgMhfStateFestaG) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfStateFestaG) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}