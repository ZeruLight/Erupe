package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUseRewardSong represents the MSG_MHF_USE_REWARD_SONG
type MsgMhfUseRewardSong struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUseRewardSong) Opcode() network.PacketID {
	return network.MSG_MHF_USE_REWARD_SONG
}

// Parse parses the packet from binary
func (m *MsgMhfUseRewardSong) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUseRewardSong) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
