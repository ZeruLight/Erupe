package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfListMember represents the MSG_MHF_LIST_MEMBER
type MsgMhfListMember struct {
	AckHandle uint32
	Unk0      uint8 // Hardcoded 01 in the JP client.
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfListMember) Opcode() network.PacketID {
	return network.MSG_MHF_LIST_MEMBER
}

// Parse parses the packet from binary
func (m *MsgMhfListMember) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	bf.ReadUint8() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfListMember) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
