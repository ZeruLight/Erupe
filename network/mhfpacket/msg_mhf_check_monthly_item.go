package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfCheckMonthlyItem represents the MSG_MHF_CHECK_MONTHLY_ITEM
type MsgMhfCheckMonthlyItem struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCheckMonthlyItem) Opcode() network.PacketID {
	return network.MSG_MHF_CHECK_MONTHLY_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfCheckMonthlyItem) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCheckMonthlyItem) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
