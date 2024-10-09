package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfEnumerateGuildItem represents the MSG_MHF_ENUMERATE_GUILD_ITEM
type MsgMhfEnumerateGuildItem struct {
	AckHandle uint32
	GuildID   uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuildItem) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUILD_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuildItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	bf.ReadUint8() // Zeroed
	bf.ReadUint8() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuildItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
