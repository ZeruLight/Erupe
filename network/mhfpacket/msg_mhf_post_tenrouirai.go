package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfPostTenrouirai represents the MSG_MHF_POST_TENROUIRAI
type MsgMhfPostTenrouirai struct {
	AckHandle uint32
	Unk0      uint8
	Unk1      uint8
	GuildID   uint32
	Unk3      uint8
	Floors    uint16
	Antiques  uint16
	Chests    uint16
	Cats      uint16
	TRP       uint16
	Slays     uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostTenrouirai) Opcode() network.PacketID {
	return network.MSG_MHF_POST_TENROUIRAI
}

// Parse parses the packet from binary
func (m *MsgMhfPostTenrouirai) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint8()
	m.GuildID = bf.ReadUint32()
	m.Unk3 = bf.ReadUint8()
	m.Floors = bf.ReadUint16()
	m.Antiques = bf.ReadUint16()
	m.Chests = bf.ReadUint16()
	m.Cats = bf.ReadUint16()
	m.TRP = bf.ReadUint16()
	m.Slays = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostTenrouirai) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
