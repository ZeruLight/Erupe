package mhfpacket

import (
	"errors"
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysOperateRegister represents the MSG_SYS_OPERATE_REGISTER
type MsgSysOperateRegister struct {
	AckHandle      uint32
	SemaphoreID    uint32
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysOperateRegister) Opcode() network.PacketID {
	return network.MSG_SYS_OPERATE_REGISTER
}

// Parse parses the packet from binary
func (m *MsgSysOperateRegister) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.SemaphoreID = bf.ReadUint32()
	_ = bf.ReadUint16()
	dataSize := bf.ReadUint16()
	m.RawDataPayload = bf.ReadBytes(uint(dataSize) - 1)
	_ = bf.ReadBytes(1) // Null terminator
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysOperateRegister) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
