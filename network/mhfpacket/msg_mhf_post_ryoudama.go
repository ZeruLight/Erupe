package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfPostRyoudama represents the MSG_MHF_POST_RYOUDAMA
type MsgMhfPostRyoudama struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostRyoudama) Opcode() network.PacketID {
	return network.MSG_MHF_POST_RYOUDAMA
}

// Parse parses the packet from binary
func (m *MsgMhfPostRyoudama) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostRyoudama) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
