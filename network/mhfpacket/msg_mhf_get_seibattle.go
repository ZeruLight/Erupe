package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetSeibattle represents the MSG_MHF_GET_SEIBATTLE
type MsgMhfGetSeibattle struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetSeibattle) Opcode() network.PacketID {
	return network.MSG_MHF_GET_SEIBATTLE
}

// Parse parses the packet from binary
func (m *MsgMhfGetSeibattle) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetSeibattle) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}