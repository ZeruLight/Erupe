package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetGemInfo represents the MSG_MHF_GET_GEM_INFO
type MsgMhfGetGemInfo struct {
	AckHandle uint32
	Unk       uint32
	Unk1      []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGemInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GEM_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetGemInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk = bf.ReadUint32()
	m.Unk1 = bf.ReadBytes(24)
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGemInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
