package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfPlayNormalGacha represents the MSG_MHF_PLAY_NORMAL_GACHA
type MsgMhfPlayNormalGacha struct {
	AckHandle uint32
	GachaID   uint32
	RollType  uint8
	GachaType uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPlayNormalGacha) Opcode() network.PacketID {
	return network.MSG_MHF_PLAY_NORMAL_GACHA
}

// Parse parses the packet from binary
func (m *MsgMhfPlayNormalGacha) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GachaID = bf.ReadUint32()
	m.RollType = bf.ReadUint8()
	m.GachaType = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPlayNormalGacha) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
