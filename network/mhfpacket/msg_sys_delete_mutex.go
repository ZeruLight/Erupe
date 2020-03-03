package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysDeleteMutex represents the MSG_SYS_DELETE_MUTEX
type MsgSysDeleteMutex struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysDeleteMutex) Opcode() network.PacketID {
	return network.MSG_SYS_DELETE_MUTEX
}

// Parse parses the packet from binary
func (m *MsgSysDeleteMutex) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysDeleteMutex) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
