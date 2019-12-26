package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysLockStage represents the MSG_SYS_LOCK_STAGE
type MsgSysLockStage struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLockStage) Opcode() network.PacketID {
	return network.MSG_SYS_LOCK_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysLockStage) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysLockStage) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}