package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReceiveCafeDurationBonus represents the MSG_MHF_RECEIVE_CAFE_DURATION_BONUS
type MsgMhfReceiveCafeDurationBonus struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReceiveCafeDurationBonus) Opcode() network.PacketID {
	return network.MSG_MHF_RECEIVE_CAFE_DURATION_BONUS
}

// Parse parses the packet from binary
func (m *MsgMhfReceiveCafeDurationBonus) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReceiveCafeDurationBonus) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}