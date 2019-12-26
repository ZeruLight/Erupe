package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdTacticsLog represents the MSG_MHF_GET_UD_TACTICS_LOG
type MsgMhfGetUdTacticsLog struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdTacticsLog) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_TACTICS_LOG
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdTacticsLog) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdTacticsLog) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}