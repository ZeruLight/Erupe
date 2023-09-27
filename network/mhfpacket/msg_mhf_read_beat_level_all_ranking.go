package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfReadBeatLevelAllRanking represents the MSG_MHF_READ_BEAT_LEVEL_ALL_RANKING
type MsgMhfReadBeatLevelAllRanking struct {
	AckHandle uint32
	Unk0      uint32
	GuildID   int32
	Unk2      int32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadBeatLevelAllRanking) Opcode() network.PacketID {
	return network.MSG_MHF_READ_BEAT_LEVEL_ALL_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfReadBeatLevelAllRanking) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.GuildID = bf.ReadInt32()
	m.Unk2 = bf.ReadInt32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadBeatLevelAllRanking) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
