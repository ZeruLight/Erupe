package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfPostCafeDurationBonusReceived represents the MSG_MHF_POST_CAFE_DURATION_BONUS_RECEIVED
type MsgMhfPostCafeDurationBonusReceived struct {
	AckHandle   uint32
	Unk0        uint32
	CafeBonusID uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostCafeDurationBonusReceived) Opcode() network.PacketID {
	return network.MSG_MHF_POST_CAFE_DURATION_BONUS_RECEIVED
}

// Parse parses the packet from binary
func (m *MsgMhfPostCafeDurationBonusReceived) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.CafeBonusID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostCafeDurationBonusReceived) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
