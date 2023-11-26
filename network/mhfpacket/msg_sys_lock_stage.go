package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysLockStage represents the MSG_SYS_LOCK_STAGE
type MsgSysLockStage struct {
	AckHandle uint32
	StageID   string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLockStage) Opcode() network.PacketID {
	return network.MSG_SYS_LOCK_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysLockStage) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	bf.ReadUint8() // Always 1
	bf.ReadUint8() // Always 1
	bf.ReadUint8() // Length StageID
	m.StageID = string(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysLockStage) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
