package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetBreakSeibatuLevelReward represents the MSG_MHF_GET_BREAK_SEIBATU_LEVEL_REWARD
type MsgMhfGetBreakSeibatuLevelReward struct {
	AckHandle uint32
	Unk0      uint32
	Unk1      int32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetBreakSeibatuLevelReward) Opcode() network.PacketID {
	return network.MSG_MHF_GET_BREAK_SEIBATU_LEVEL_REWARD
}

// Parse parses the packet from binary
func (m *MsgMhfGetBreakSeibatuLevelReward) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadInt32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetBreakSeibatuLevelReward) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
