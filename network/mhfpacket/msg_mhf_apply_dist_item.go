package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfApplyDistItem represents the MSG_MHF_APPLY_DIST_ITEM
type MsgMhfApplyDistItem struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfApplyDistItem) Opcode() network.PacketID {
	return network.MSG_MHF_APPLY_DIST_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfApplyDistItem) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfApplyDistItem) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
