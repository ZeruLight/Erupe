package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateUnionItem represents the MSG_MHF_UPDATE_UNION_ITEM
type MsgMhfUpdateUnionItem struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateUnionItem) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_UNION_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateUnionItem) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateUnionItem) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
