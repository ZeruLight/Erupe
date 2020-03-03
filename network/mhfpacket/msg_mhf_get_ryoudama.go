package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetRyoudama represents the MSG_MHF_GET_RYOUDAMA
type MsgMhfGetRyoudama struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetRyoudama) Opcode() network.PacketID {
	return network.MSG_MHF_GET_RYOUDAMA
}

// Parse parses the packet from binary
func (m *MsgMhfGetRyoudama) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetRyoudama) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
