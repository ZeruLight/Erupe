package mhfpacket

import (
	"errors"
	"fmt"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetWeeklySeibatuRankingReward represents the MSG_MHF_GET_WEEKLY_SEIBATU_RANKING_REWARD
type MsgMhfGetWeeklySeibatuRankingReward struct {
	AckHandle    uint32
	Unk0         uint32
	Operation    uint32
	ID           uint32
	EarthMonster uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetWeeklySeibatuRankingReward) Opcode() network.PacketID {
	return network.MSG_MHF_GET_WEEKLY_SEIBATU_RANKING_REWARD
}

// Parse parses the packet from binary
func (m *MsgMhfGetWeeklySeibatuRankingReward) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Operation = bf.ReadUint32()
	m.ID = bf.ReadUint32()
	m.EarthMonster = bf.ReadUint32()
	fmt.Printf("MsgMhfGetWeeklySeibatuRankingReward: Unk0:[%d] Operation:[%d] ID:[%d] EarthMonster:[%d]\n\n", m.Unk0, m.Operation, m.ID, m.EarthMonster)

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetWeeklySeibatuRankingReward) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
