package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAddUdTacticsPoint represents the MSG_MHF_ADD_UD_TACTICS_POINT
type MsgMhfAddUdTacticsPoint struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddUdTacticsPoint) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_UD_TACTICS_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfAddUdTacticsPoint) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddUdTacticsPoint) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
