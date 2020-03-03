package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCreateSemaphore represents the MSG_SYS_CREATE_SEMAPHORE
type MsgSysCreateSemaphore struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCreateSemaphore) Opcode() network.PacketID {
	return network.MSG_SYS_CREATE_SEMAPHORE
}

// Parse parses the packet from binary
func (m *MsgSysCreateSemaphore) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysCreateSemaphore) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
