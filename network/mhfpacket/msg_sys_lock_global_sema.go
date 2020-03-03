package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysLockGlobalSema represents the MSG_SYS_LOCK_GLOBAL_SEMA
type MsgSysLockGlobalSema struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLockGlobalSema) Opcode() network.PacketID {
	return network.MSG_SYS_LOCK_GLOBAL_SEMA
}

// Parse parses the packet from binary
func (m *MsgSysLockGlobalSema) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysLockGlobalSema) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
