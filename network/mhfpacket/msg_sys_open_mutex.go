package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysOpenMutex represents the MSG_SYS_OPEN_MUTEX
type MsgSysOpenMutex struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysOpenMutex) Opcode() network.PacketID {
	return network.MSG_SYS_OPEN_MUTEX
}

// Parse parses the packet from binary
func (m *MsgSysOpenMutex) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysOpenMutex) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
