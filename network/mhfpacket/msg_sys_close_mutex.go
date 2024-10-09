package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysCloseMutex represents the MSG_SYS_CLOSE_MUTEX
type MsgSysCloseMutex struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCloseMutex) Opcode() network.PacketID {
	return network.MSG_SYS_CLOSE_MUTEX
}

// Parse parses the packet from binary
func (m *MsgSysCloseMutex) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysCloseMutex) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
