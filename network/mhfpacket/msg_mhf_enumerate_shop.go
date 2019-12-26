package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateShop represents the MSG_MHF_ENUMERATE_SHOP
type MsgMhfEnumerateShop struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateShop) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_SHOP
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateShop) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateShop) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}