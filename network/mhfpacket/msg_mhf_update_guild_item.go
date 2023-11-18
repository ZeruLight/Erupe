package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

type Item struct {
	Unk0   uint32
	ItemID uint16
	Amount uint16
	Unk1   uint32
}

// MsgMhfUpdateGuildItem represents the MSG_MHF_UPDATE_GUILD_ITEM
type MsgMhfUpdateGuildItem struct {
	AckHandle uint32
	GuildId   uint32
	Items     []Item
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateGuildItem) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_GUILD_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateGuildItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GuildId = bf.ReadUint32()
	itemCount := int(bf.ReadUint16())
	bf.ReadUint8() // Zeroed
	bf.ReadUint8() // Zeroed
	m.Items = make([]Item, itemCount)

	for i := 0; i < itemCount; i++ {
		m.Items[i].Unk0 = bf.ReadUint32()
		m.Items[i].ItemID = bf.ReadUint16()
		m.Items[i].Amount = bf.ReadUint16()
		m.Items[i].Unk1 = bf.ReadUint32()
	}

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateGuildItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
