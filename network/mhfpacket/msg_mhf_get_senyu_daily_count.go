package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetSenyuDailyCount represents the MSG_MHF_GET_SENYU_DAILY_COUNT
type MsgMhfGetSenyuDailyCount struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetSenyuDailyCount) Opcode() network.PacketID {
	return network.MSG_MHF_GET_SENYU_DAILY_COUNT
}

// Parse parses the packet from binary
func (m *MsgMhfGetSenyuDailyCount) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetSenyuDailyCount) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
