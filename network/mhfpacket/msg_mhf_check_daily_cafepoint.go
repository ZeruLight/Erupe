package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfCheckDailyCafepoint represents the MSG_MHF_CHECK_DAILY_CAFEPOINT
type MsgMhfCheckDailyCafepoint struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCheckDailyCafepoint) Opcode() network.PacketID {
	return network.MSG_MHF_CHECK_DAILY_CAFEPOINT
}

// Parse parses the packet from binary
func (m *MsgMhfCheckDailyCafepoint) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCheckDailyCafepoint) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
