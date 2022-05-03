package mhfpacket

import (
 "errors"

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

type OperateJointAction uint8

const (
  OPERATE_JOINT_DISBAND = 0x01
  OPERATE_JOINT_LEAVE = 0x03
  OPERATE_JOINT_KICK = 0x09
)

// MsgMhfOperateJoint represents the MSG_MHF_OPERATE_JOINT
type MsgMhfOperateJoint struct {
  AckHandle uint32
  AllianceID uint32
  GuildID uint32
  Action OperateJointAction
  UnkData []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfOperateJoint) Opcode() network.PacketID {
	return network.MSG_MHF_OPERATE_JOINT
}

// Parse parses the packet from binary
func (m *MsgMhfOperateJoint) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
  m.AckHandle = bf.ReadUint32()
  m.AllianceID = bf.ReadUint32()
  m.GuildID = bf.ReadUint32()
  m.Action = OperateJointAction(bf.ReadUint8())
  m.UnkData = bf.DataFromCurrent()
  bf.Seek(int64(len(bf.Data()) - 2), 0)
  return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOperateJoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
