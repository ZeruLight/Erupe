package mhfpacket

import (
	"errors"
	"erupe-ce/common/mhfitem"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfUpdateGuildItem represents the MSG_MHF_UPDATE_GUILD_ITEM
type MsgMhfUpdateGuildItem struct {
	AckHandle    uint32
	GuildID      uint32
	UpdatedItems []mhfitem.MHFItemStack
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateGuildItem) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_GUILD_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateGuildItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	changes := int(bf.ReadUint16())
	bf.ReadUint8() // Zeroed
	bf.ReadUint8() // Zeroed
	for i := 0; i < changes; i++ {
		m.UpdatedItems = append(m.UpdatedItems, mhfitem.ReadWarehouseItem(bf))
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateGuildItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
