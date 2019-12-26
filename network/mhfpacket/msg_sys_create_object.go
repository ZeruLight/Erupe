package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCreateObject represents the MSG_SYS_CREATE_OBJECT
type MsgSysCreateObject struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCreateObject) Opcode() network.PacketID {
	return network.MSG_SYS_CREATE_OBJECT
}

// Parse parses the packet from binary
func (m *MsgSysCreateObject) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysCreateObject) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}