package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetNotice represents the MSG_MHF_GET_NOTICE
type MsgMhfGetNotice struct {
	AckHandle uint32
	Unk0      uint32
	Unk1      uint32
	Unk2      int32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetNotice) Opcode() network.PacketID {
	return network.MSG_MHF_GET_NOTICE
}

// Parse parses the packet from binary
func (m *MsgMhfGetNotice) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadInt32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetNotice) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
