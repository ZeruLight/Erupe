package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetTinyBin represents the MSG_MHF_GET_TINY_BIN
type MsgMhfGetTinyBin struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetTinyBin) Opcode() network.PacketID {
	return network.MSG_MHF_GET_TINY_BIN
}

// Parse parses the packet from binary
func (m *MsgMhfGetTinyBin) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetTinyBin) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}