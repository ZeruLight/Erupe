package mhfpacket

import (
 "errors"

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateUnionItem represents the MSG_MHF_UPDATE_UNION_ITEM
type MsgMhfUpdateUnionItem struct {
	AckHandle uint32
	Amount    uint16
	Unk1      uint16 // 0x00 0x00
	Items     []Item // Array of updated item IDs
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateUnionItem) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_UNION_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateUnionItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Amount = bf.ReadUint16()
	m.Unk1 = bf.ReadUint16()
	m.Items = make([]Item, int(m.Amount))

	for i := 0; i < int(m.Amount); i++ {
		m.Items[i].Unk0 = bf.ReadUint32()
		m.Items[i].ItemId = bf.ReadUint16()
		m.Items[i].Amount = bf.ReadUint16()
		m.Items[i].Unk1 = bf.ReadUint32()
	}

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateUnionItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
