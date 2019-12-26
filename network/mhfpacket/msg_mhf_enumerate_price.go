package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumeratePrice represents the MSG_MHF_ENUMERATE_PRICE
type MsgMhfEnumeratePrice struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumeratePrice) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_PRICE
}

// Parse parses the packet from binary
func (m *MsgMhfEnumeratePrice) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumeratePrice) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}