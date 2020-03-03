package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetBreakSeibatuLevelReward represents the MSG_MHF_GET_BREAK_SEIBATU_LEVEL_REWARD
type MsgMhfGetBreakSeibatuLevelReward struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetBreakSeibatuLevelReward) Opcode() network.PacketID {
	return network.MSG_MHF_GET_BREAK_SEIBATU_LEVEL_REWARD
}

// Parse parses the packet from binary
func (m *MsgMhfGetBreakSeibatuLevelReward) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetBreakSeibatuLevelReward) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
