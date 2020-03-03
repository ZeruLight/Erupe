package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysAcquireSemaphore represents the MSG_SYS_ACQUIRE_SEMAPHORE
type MsgSysAcquireSemaphore struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysAcquireSemaphore) Opcode() network.PacketID {
	return network.MSG_SYS_ACQUIRE_SEMAPHORE
}

// Parse parses the packet from binary
func (m *MsgSysAcquireSemaphore) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysAcquireSemaphore) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
