package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetRewardSong represents the MSG_MHF_GET_REWARD_SONG
type MsgMhfGetRewardSong struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetRewardSong) Opcode() network.PacketID {
	return network.MSG_MHF_GET_REWARD_SONG
}

// Parse parses the packet from binary
func (m *MsgMhfGetRewardSong) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetRewardSong) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
