package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCleanupObject represents the MSG_SYS_CLEANUP_OBJECT
type MsgSysCleanupObject struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCleanupObject) Opcode() network.PacketID {
	return network.MSG_SYS_CLEANUP_OBJECT
}

// Parse parses the packet from binary
func (m *MsgSysCleanupObject) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysCleanupObject) Build(bf *byteframe.ByteFrame) error {
	// This packet has no data.
	return nil
}
