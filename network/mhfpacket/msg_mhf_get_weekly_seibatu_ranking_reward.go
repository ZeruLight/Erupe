package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetWeeklySeibatuRankingReward represents the MSG_MHF_GET_WEEKLY_SEIBATU_RANKING_REWARD
type MsgMhfGetWeeklySeibatuRankingReward struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetWeeklySeibatuRankingReward) Opcode() network.PacketID {
	return network.MSG_MHF_GET_WEEKLY_SEIBATU_RANKING_REWARD
}

// Parse parses the packet from binary
func (m *MsgMhfGetWeeklySeibatuRankingReward) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetWeeklySeibatuRankingReward) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
