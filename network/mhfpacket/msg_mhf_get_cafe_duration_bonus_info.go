package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetCafeDurationBonusInfo represents the MSG_MHF_GET_CAFE_DURATION_BONUS_INFO
type MsgMhfGetCafeDurationBonusInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetCafeDurationBonusInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_CAFE_DURATION_BONUS_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetCafeDurationBonusInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetCafeDurationBonusInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}