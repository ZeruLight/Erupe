package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAddRewardSongCount represents the MSG_MHF_ADD_REWARD_SONG_COUNT
type MsgMhfAddRewardSongCount struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddRewardSongCount) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_REWARD_SONG_COUNT
}

// Parse parses the packet from binary
func (m *MsgMhfAddRewardSongCount) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddRewardSongCount) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
