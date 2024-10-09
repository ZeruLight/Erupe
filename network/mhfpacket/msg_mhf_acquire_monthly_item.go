package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfAcquireMonthlyItem represents the MSG_MHF_ACQUIRE_MONTHLY_ITEM
type MsgMhfAcquireMonthlyItem struct {
	AckHandle uint32
	Unk0      uint8
	Unk1      uint8
	Unk2      uint16
	Unk3      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireMonthlyItem) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_MONTHLY_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireMonthlyItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint8()
	m.Unk2 = bf.ReadUint16()
	m.Unk3 = bf.ReadUint32()
	bf.ReadUint32() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireMonthlyItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
