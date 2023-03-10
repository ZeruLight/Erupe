package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetRyoudama represents the MSG_MHF_GET_RYOUDAMA
type MsgMhfGetRyoudama struct {
	AckHandle uint32
	Unk0      uint8
	Unk1      uint8
	GuildID   uint32
	Unk3      uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetRyoudama) Opcode() network.PacketID {
	return network.MSG_MHF_GET_RYOUDAMA
}

// Parse parses the packet from binary
func (m *MsgMhfGetRyoudama) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint8()
	m.GuildID = bf.ReadUint32()
	m.Unk3 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetRyoudama) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
