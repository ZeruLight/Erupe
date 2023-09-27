package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfSetGuildManageRight represents the MSG_MHF_SET_GUILD_MANAGE_RIGHT
type MsgMhfSetGuildManageRight struct {
	AckHandle uint32
	CharID    uint32
	Allowed   bool
	Unk       []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetGuildManageRight) Opcode() network.PacketID {
	return network.MSG_MHF_SET_GUILD_MANAGE_RIGHT
}

// Parse parses the packet from binary
func (m *MsgMhfSetGuildManageRight) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.CharID = bf.ReadUint32()
	m.Allowed = bf.ReadBool()
	m.Unk = bf.ReadBytes(3)
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetGuildManageRight) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
