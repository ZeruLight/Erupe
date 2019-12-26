package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetBoostRight represents the MSG_MHF_GET_BOOST_RIGHT
type MsgMhfGetBoostRight struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetBoostRight) Opcode() network.PacketID {
	return network.MSG_MHF_GET_BOOST_RIGHT
}

// Parse parses the packet from binary
func (m *MsgMhfGetBoostRight) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetBoostRight) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}