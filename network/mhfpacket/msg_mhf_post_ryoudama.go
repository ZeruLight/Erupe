package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfPostRyoudama represents the MSG_MHF_POST_RYOUDAMA
type MsgMhfPostRyoudama struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostRyoudama) Opcode() network.PacketID {
	return network.MSG_MHF_POST_RYOUDAMA
}

// Parse parses the packet from binary
func (m *MsgMhfPostRyoudama) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostRyoudama) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
