package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfPostCafeDurationBonusReceived represents the MSG_MHF_POST_CAFE_DURATION_BONUS_RECEIVED
type MsgMhfPostCafeDurationBonusReceived struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostCafeDurationBonusReceived) Opcode() network.PacketID {
	return network.MSG_MHF_POST_CAFE_DURATION_BONUS_RECEIVED
}

// Parse parses the packet from binary
func (m *MsgMhfPostCafeDurationBonusReceived) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostCafeDurationBonusReceived) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
