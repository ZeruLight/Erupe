package mhfpacket

import (
	"errors"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfEnumerateGuildMember represents the MSG_MHF_ENUMERATE_GUILD_MEMBER
type MsgMhfEnumerateGuildMember struct {
	AckHandle  uint32
	AllianceID uint32
	GuildID    uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuildMember) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUILD_MEMBER
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuildMember) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	bf.ReadUint8() // Zeroed
	bf.ReadUint8() // Always 1
	m.AllianceID = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuildMember) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
