package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfAddRewardSongCount represents the MSG_MHF_ADD_REWARD_SONG_COUNT
type MsgMhfAddRewardSongCount struct {
	AckHandle uint32
	Unk0      uint32
	Unk1      []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddRewardSongCount) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_REWARD_SONG_COUNT
}

// Parse parses the packet from binary
func (m *MsgMhfAddRewardSongCount) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadBytes(5)
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddRewardSongCount) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
