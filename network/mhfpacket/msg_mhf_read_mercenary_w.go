package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfReadMercenaryW represents the MSG_MHF_READ_MERCENARY_W
type MsgMhfReadMercenaryW struct {
	AckHandle uint32
	Op        uint8
	Unk1      uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadMercenaryW) Opcode() network.PacketID {
	return network.MSG_MHF_READ_MERCENARY_W
}

// Parse parses the packet from binary
func (m *MsgMhfReadMercenaryW) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Op = bf.ReadUint8()
	m.Unk1 = bf.ReadUint8() // Supposed to be 0 or 1, but always 1
	bf.ReadUint8()          // Zeroed
	bf.ReadUint8()          // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadMercenaryW) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
