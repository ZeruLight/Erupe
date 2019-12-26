package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdTacticsPoint represents the MSG_MHF_GET_UD_TACTICS_POINT
type MsgMhfGetUdTacticsPoint struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdTacticsPoint) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_TACTICS_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdTacticsPoint) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdTacticsPoint) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}