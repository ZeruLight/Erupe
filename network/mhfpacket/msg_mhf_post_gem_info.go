package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfPostGemInfo represents the MSG_MHF_POST_GEM_INFO
type MsgMhfPostGemInfo struct {
	AckHandle uint32
	Op        uint32
	Unk1      uint32
	Gem       int32
	Quantity  int32
	CID       int32
	Message   int32
	Unk6      int32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostGemInfo) Opcode() network.PacketID {
	return network.MSG_MHF_POST_GEM_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfPostGemInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Op = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Gem = bf.ReadInt32()
	m.Quantity = bf.ReadInt32()
	m.CID = bf.ReadInt32()
	m.Message = bf.ReadInt32()
	m.Unk6 = bf.ReadInt32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostGemInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
