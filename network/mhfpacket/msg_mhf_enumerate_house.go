package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateHouse represents the MSG_MHF_ENUMERATE_HOUSE
type MsgMhfEnumerateHouse struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateHouse) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_HOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateHouse) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateHouse) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
