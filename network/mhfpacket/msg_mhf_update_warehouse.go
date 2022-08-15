package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

type UpdatedStack struct {
	ID       uint32
	Index    uint16
	ItemID   uint16
	Quantity uint16
	Unk      uint16
}

// MsgMhfUpdateWarehouse represents the MSG_MHF_UPDATE_WAREHOUSE
type MsgMhfUpdateWarehouse struct {
	AckHandle uint32
	BoxID     uint16
	Items     []UpdatedStack
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateWarehouse) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_WAREHOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateWarehouse) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.BoxID = bf.ReadUint16()
	changes := int(bf.ReadUint16())
	var stackUpdate UpdatedStack
	for i := 0; i < changes; i++ {
		stackUpdate.ID = bf.ReadUint32()
		stackUpdate.Index = bf.ReadUint16()
		stackUpdate.ItemID = bf.ReadUint16()
		stackUpdate.Quantity = bf.ReadUint16()
		stackUpdate.Unk = bf.ReadUint16()
		m.Items = append(m.Items, stackUpdate)
	}
	_ = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateWarehouse) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
