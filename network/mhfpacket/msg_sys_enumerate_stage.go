package mhfpacket

import (
	"errors"
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysEnumerateStage represents the MSG_SYS_ENUMERATE_STAGE
type MsgSysEnumerateStage struct {
	AckHandle   uint32
	StagePrefix string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysEnumerateStage) Opcode() network.PacketID {
	return network.MSG_SYS_ENUMERATE_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysEnumerateStage) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	bf.ReadUint8() // Always 1
	bf.ReadUint8() // Length StagePrefix
	m.StagePrefix = string(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysEnumerateStage) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
