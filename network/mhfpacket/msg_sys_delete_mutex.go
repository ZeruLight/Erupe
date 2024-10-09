package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysDeleteMutex represents the MSG_SYS_DELETE_MUTEX
type MsgSysDeleteMutex struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysDeleteMutex) Opcode() network.PacketID {
	return network.MSG_SYS_DELETE_MUTEX
}

// Parse parses the packet from binary
func (m *MsgSysDeleteMutex) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysDeleteMutex) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
