package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdRankingRewardList represents the MSG_MHF_GET_UD_RANKING_REWARD_LIST
type MsgMhfGetUdRankingRewardList struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdRankingRewardList) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_RANKING_REWARD_LIST
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdRankingRewardList) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdRankingRewardList) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}