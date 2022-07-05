package mhfpacket

import (
 "errors"

 	"erupe-ce/common/bfutil"
 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfCreateJoint represents the MSG_MHF_CREATE_JOINT
type MsgMhfCreateJoint struct {
  AckHandle uint32
  GuildID uint32
  Name string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCreateJoint) Opcode() network.PacketID {
	return network.MSG_MHF_CREATE_JOINT
}

// Parse parses the packet from binary
func (m *MsgMhfCreateJoint) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
  m.AckHandle = bf.ReadUint32()
  m.GuildID = bf.ReadUint32()
  nameLength := bf.ReadUint32()
  nameBytes := bfutil.UpToNull(bf.ReadBytes(uint(nameLength)))
  m.Name = ctx.StrConv.MustDecode(nameBytes)
  return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCreateJoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
