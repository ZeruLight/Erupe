package mhfpacket

import (
	"errors"
	"erupe-ce/utils/bfutil"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgSysAcquireSemaphore represents the MSG_SYS_ACQUIRE_SEMAPHORE
type MsgSysAcquireSemaphore struct {
	AckHandle   uint32
	SemaphoreID string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysAcquireSemaphore) Opcode() network.PacketID {
	return network.MSG_SYS_ACQUIRE_SEMAPHORE
}

// Parse parses the packet from binary
func (m *MsgSysAcquireSemaphore) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	SemaphoreIDLength := bf.ReadUint8()
	m.SemaphoreID = string(bfutil.UpToNull(bf.ReadBytes(uint(SemaphoreIDLength))))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysAcquireSemaphore) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
