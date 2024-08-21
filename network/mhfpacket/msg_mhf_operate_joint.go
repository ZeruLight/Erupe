package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

type OperateJointAction uint8

const (
	OperateJointDisband = 1
	OperateJointLeave   = 3
	OperateJointKick    = 9
)

// MsgMhfOperateJoint represents the MSG_MHF_OPERATE_JOINT
type MsgMhfOperateJoint struct {
	AckHandle  uint32
	AllianceID uint32
	GuildID    uint32
	Action     OperateJointAction
	Data1      *byteframe.ByteFrame
	Data2      *byteframe.ByteFrame
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
	dataLen := uint(bf.ReadUint8())
	m.Data1 = byteframe.NewByteFrameFromBytes(bf.ReadBytes(4))
	m.Data2 = byteframe.NewByteFrameFromBytes(bf.ReadBytes(dataLen))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOperateJoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
