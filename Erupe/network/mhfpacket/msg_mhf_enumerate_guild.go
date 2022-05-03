package mhfpacket

import (
 "errors"

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

type EnumerateGuildType uint8

const (
  ENUMERATE_GUILD_TYPE_GUILD_NAME = 0x01
  ENUMERATE_GUILD_TYPE_LEADER_NAME = 0x02
  ENUMERATE_GUILD_TYPE_LEADER_ID = 0x03
  ENUMERATE_GUILD_TYPE_ORDER_MEMBERS = 0x04
  ENUMERATE_GUILD_TYPE_ORDER_REGISTRATION = 0x05
  ENUMERATE_GUILD_TYPE_ORDER_RANK = 0x06
  ENUMERATE_GUILD_TYPE_MOTTO = 0x07
  ENUMERATE_GUILD_TYPE_RECRUITING = 0x08
  ENUMERATE_ALLIANCE_TYPE_ALLIANCE_NAME = 0x09
  ENUMERATE_ALLIANCE_TYPE_LEADER_NAME = 0x0A
  ENUMERATE_ALLIANCE_TYPE_LEADER_ID = 0x0B
  ENUMERATE_ALLIANCE_TYPE_ORDER_MEMBERS = 0x0C
  ENUMERATE_ALLIANCE_TYPE_ORDER_REGISTRATION = 0x0D
)

// MsgMhfEnumerateGuild represents the MSG_MHF_ENUMERATE_GUILD
type MsgMhfEnumerateGuild struct {
	AckHandle      uint32
	Type           EnumerateGuildType
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuild) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuild) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Type = EnumerateGuildType(bf.ReadUint8())
	m.RawDataPayload = bf.DataFromCurrent()
  bf.Seek(int64(len(bf.Data()) - 2), 0)
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuild) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
