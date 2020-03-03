package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCheckSemaphore represents the MSG_SYS_CHECK_SEMAPHORE
type MsgSysCheckSemaphore struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCheckSemaphore) Opcode() network.PacketID {
	return network.MSG_SYS_CHECK_SEMAPHORE
}

// Parse parses the packet from binary
func (m *MsgSysCheckSemaphore) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysCheckSemaphore) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
