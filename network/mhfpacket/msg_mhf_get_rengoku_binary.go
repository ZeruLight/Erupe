package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetRengokuBinary represents the MSG_MHF_GET_RENGOKU_BINARY
type MsgMhfGetRengokuBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetRengokuBinary) Opcode() network.PacketID {
	return network.MSG_MHF_GET_RENGOKU_BINARY
}

// Parse parses the packet from binary
func (m *MsgMhfGetRengokuBinary) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetRengokuBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}