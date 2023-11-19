package mhfpacket

import (
	"errors"
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfArrangeGuildMember represents the MSG_MHF_ARRANGE_GUILD_MEMBER
type MsgMhfArrangeGuildMember struct {
	AckHandle uint32
	GuildID   uint32
	CharIDs   []uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfArrangeGuildMember) Opcode() network.PacketID {
	return network.MSG_MHF_ARRANGE_GUILD_MEMBER
}

// Parse parses the packet from binary
func (m *MsgMhfArrangeGuildMember) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	bf.ReadUint8() // Zeroed
	charCount := int(bf.ReadUint8())
	m.CharIDs = make([]uint32, charCount)

	for i := 0; i < charCount; i++ {
		m.CharIDs[i] = bf.ReadUint32()
	}

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfArrangeGuildMember) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
