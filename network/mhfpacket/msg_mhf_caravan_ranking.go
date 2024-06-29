package mhfpacket

import (
	"errors"
	"fmt"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfCaravanRanking represents the MSG_MHF_CARAVAN_RANKING
type MsgMhfCaravanRanking struct {
	AckHandle     uint32
	Unk0          uint32
	Operation     uint32
	HunterGroupId int32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCaravanRanking) Opcode() network.PacketID {
	return network.MSG_MHF_CARAVAN_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfCaravanRanking) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Operation = bf.ReadUint32()
	m.HunterGroupId = bf.ReadInt32()
	fmt.Printf("Unk0:[%d] Operation:[%d] HunterGroupId:[%d] \n\n", m.Unk0, m.Operation, m.HunterGroupId)

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCaravanRanking) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
