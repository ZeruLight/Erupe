package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdDailyPresentList represents the MSG_MHF_GET_UD_DAILY_PRESENT_LIST
type MsgMhfGetUdDailyPresentList struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdDailyPresentList) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_DAILY_PRESENT_LIST
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdDailyPresentList) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdDailyPresentList) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}