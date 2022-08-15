package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfEnumerateWarehouse represents the MSG_MHF_ENUMERATE_WAREHOUSE
type MsgMhfEnumerateWarehouse struct {
	AckHandle uint32
	BoxType   string
	BoxIndex  uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateWarehouse) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_WAREHOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateWarehouse) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	boxType := bf.ReadUint8()
	switch boxType {
	case 0:
		m.BoxType = "item"
	case 1:
		m.BoxType = "equip"
	}
	m.BoxIndex = bf.ReadUint8()
	_ = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateWarehouse) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
