package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdTacticsRewardList represents the MSG_MHF_GET_UD_TACTICS_REWARD_LIST
type MsgMhfGetUdTacticsRewardList struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdTacticsRewardList) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_TACTICS_REWARD_LIST
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdTacticsRewardList) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdTacticsRewardList) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}