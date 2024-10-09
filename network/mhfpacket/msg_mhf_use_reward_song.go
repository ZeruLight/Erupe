package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
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
	return errors.New("NOT IMPLEMENTED")
}
