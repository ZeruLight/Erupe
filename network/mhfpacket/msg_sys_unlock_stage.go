package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysUnlockStage represents the MSG_SYS_UNLOCK_STAGE
type MsgSysUnlockStage struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysUnlockStage) Opcode() network.PacketID {
	return network.MSG_SYS_UNLOCK_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysUnlockStage) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysUnlockStage) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}