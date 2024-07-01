package mhfpacket

import (
	"errors"
	"fmt"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetFixedSeibatuRankingTable represents the MSG_MHF_GET_FIXED_SEIBATU_RANKING_TABLE
type MsgMhfGetFixedSeibatuRankingTable struct {
	AckHandle    uint32
	Unk0         uint32
	Unk1         int32
	EarthMonster int32
	Unk3         int32
	Unk4         int32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetFixedSeibatuRankingTable) Opcode() network.PacketID {
	return network.MSG_MHF_GET_FIXED_SEIBATU_RANKING_TABLE
}

// Parse parses the packet from binary
func (m *MsgMhfGetFixedSeibatuRankingTable) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadInt32()
	m.EarthMonster = bf.ReadInt32()
	m.Unk3 = bf.ReadInt32()
	m.Unk4 = bf.ReadInt32()
	fmt.Printf("MsgMhfGetFixedSeibatuRankingTable: Unk0:[%d] Unk1:[%d] EarthMonster:[%d] Unk3:[%d] Unk4:[%d]\n\n", m.Unk0, m.Unk1, m.EarthMonster, m.Unk3, m.Unk4)

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetFixedSeibatuRankingTable) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
