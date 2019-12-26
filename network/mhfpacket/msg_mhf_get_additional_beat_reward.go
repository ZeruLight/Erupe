package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetAdditionalBeatReward represents the MSG_MHF_GET_ADDITIONAL_BEAT_REWARD
type MsgMhfGetAdditionalBeatReward struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetAdditionalBeatReward) Opcode() network.PacketID {
	return network.MSG_MHF_GET_ADDITIONAL_BEAT_REWARD
}

// Parse parses the packet from binary
func (m *MsgMhfGetAdditionalBeatReward) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetAdditionalBeatReward) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}