package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysDuplicateObject represents the MSG_SYS_DUPLICATE_OBJECT
type MsgSysDuplicateObject struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysDuplicateObject) Opcode() network.PacketID {
	return network.MSG_SYS_DUPLICATE_OBJECT
}

// Parse parses the packet from binary
func (m *MsgSysDuplicateObject) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysDuplicateObject) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}