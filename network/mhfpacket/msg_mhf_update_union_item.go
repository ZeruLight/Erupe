package mhfpacket

import (
	"errors"
	"erupe-ce/utils/mhfitem"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfUpdateUnionItem represents the MSG_MHF_UPDATE_UNION_ITEM
type MsgMhfUpdateUnionItem struct {
	AckHandle    uint32
	UpdatedItems []mhfitem.MHFItemStack
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateUnionItem) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_UNION_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateUnionItem) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	changes := int(bf.ReadUint16())
	bf.ReadUint8() // Zeroed
	bf.ReadUint8() // Zeroed
	for i := 0; i < changes; i++ {
		m.UpdatedItems = append(m.UpdatedItems, mhfitem.ReadWarehouseItem(bf))
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateUnionItem) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
