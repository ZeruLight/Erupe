package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcquireMonthlyItem represents the MSG_MHF_ACQUIRE_MONTHLY_ITEM
type MsgMhfAcquireMonthlyItem struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireMonthlyItem) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_MONTHLY_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireMonthlyItem) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireMonthlyItem) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
