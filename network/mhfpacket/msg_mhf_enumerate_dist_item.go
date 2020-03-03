package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateDistItem represents the MSG_MHF_ENUMERATE_DIST_ITEM
type MsgMhfEnumerateDistItem struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateDistItem) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_DIST_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateDistItem) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateDistItem) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
