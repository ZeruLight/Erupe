package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCreateOpenMutex represents the MSG_SYS_CREATE_OPEN_MUTEX
type MsgSysCreateOpenMutex struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCreateOpenMutex) Opcode() network.PacketID {
	return network.MSG_SYS_CREATE_OPEN_MUTEX
}

// Parse parses the packet from binary
func (m *MsgSysCreateOpenMutex) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysCreateOpenMutex) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
