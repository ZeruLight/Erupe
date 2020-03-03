package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReleaseSemaphore represents the MSG_SYS_RELEASE_SEMAPHORE
type MsgSysReleaseSemaphore struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReleaseSemaphore) Opcode() network.PacketID {
	return network.MSG_SYS_RELEASE_SEMAPHORE
}

// Parse parses the packet from binary
func (m *MsgSysReleaseSemaphore) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReleaseSemaphore) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
