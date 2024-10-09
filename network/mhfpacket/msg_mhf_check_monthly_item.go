package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfCheckMonthlyItem represents the MSG_MHF_CHECK_MONTHLY_ITEM
type MsgMhfCheckMonthlyItem struct {
	AckHandle uint32
	Type      uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCheckMonthlyItem) Opcode() network.PacketID {
	return network.MSG_MHF_CHECK_MONTHLY_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfCheckMonthlyItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Type = bf.ReadUint8()
	bf.ReadUint8() // Zeroed
	bf.ReadUint8() // Zeroed
	bf.ReadUint8() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCheckMonthlyItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
