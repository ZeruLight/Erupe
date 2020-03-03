package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdTacticsRemainingPoint represents the MSG_MHF_GET_UD_TACTICS_REMAINING_POINT
type MsgMhfGetUdTacticsRemainingPoint struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdTacticsRemainingPoint) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_TACTICS_REMAINING_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdTacticsRemainingPoint) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdTacticsRemainingPoint) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
