package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfReadBeatLevelMyRanking represents the MSG_MHF_READ_BEAT_LEVEL_MY_RANKING
type MsgMhfReadBeatLevelMyRanking struct {
	AckHandle uint32
	Unk0      uint32
	Unk1      uint32
	Unk2      []int32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadBeatLevelMyRanking) Opcode() network.PacketID {
	return network.MSG_MHF_READ_BEAT_LEVEL_MY_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfReadBeatLevelMyRanking) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	for i := 0; i < 16; i++ {
		m.Unk2 = append(m.Unk2, bf.ReadInt32())
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadBeatLevelMyRanking) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
