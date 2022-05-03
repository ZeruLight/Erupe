package mhfpacket

import (
 "errors"

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

type OperateGuildAction uint8

const (
  OPERATE_GUILD_DISBAND = 0x01
  OPERATE_GUILD_APPLY = 0x02
  OPERATE_GUILD_LEAVE = 0x03
  OPERATE_GUILD_RESIGN = 0x04
  OPERATE_GUILD_SET_APPLICATION_DENY = 0x05
  OPERATE_GUILD_SET_APPLICATION_ALLOW = 0x06
  OPERATE_GUILD_SET_AVOID_LEADERSHIP_TRUE = 0x07
  OPERATE_GUILD_SET_AVOID_LEADERSHIP_FALSE = 0x08
  OPERATE_GUILD_UPDATE_COMMENT = 0x09
  OPERATE_GUILD_DONATE_RANK = 0x0a
  OPERATE_GUILD_UPDATE_MOTTO = 0x0b
  OPERATE_GUILD_RENAME_PUGI_1 = 0x0c
  OPERATE_GUILD_RENAME_PUGI_2 = 0x0d
  OPERATE_GUILD_RENAME_PUGI_3 = 0x0e
  OPERATE_GUILD_CHANGE_PUGI_1 = 0x0f
  OPERATE_GUILD_CHANGE_PUGI_2 = 0x10
  OPERATE_GUILD_CHANGE_PUGI_3 = 0x11
  // pugi something
  OPERATE_GUILD_DONATE_EVENT = 0x15
  // pugi something
  OPERATE_GUILD_CHANGE_DIVA_PUGI_1 = 0x19
  OPERATE_GUILD_CHANGE_DIVA_PUGI_2 = 0x1a
  OPERATE_GUILD_CHANGE_DIVA_PUGI_3 = 0x1b
)

// MsgMhfOperateGuild represents the MSG_MHF_OPERATE_GUILD
type MsgMhfOperateGuild struct {
	AckHandle uint32
	GuildID   uint32
	Action    OperateGuildAction
	UnkData   []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfOperateGuild) Opcode() network.PacketID {
	return network.MSG_MHF_OPERATE_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfOperateGuild) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	m.Action = OperateGuildAction(bf.ReadUint8())
	m.UnkData = bf.DataFromCurrent()
  bf.Seek(int64(len(bf.Data()) - 2), 0)
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOperateGuild) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
