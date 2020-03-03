package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgHead represents the MSG_HEAD
type MsgHead struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgHead) Opcode() network.PacketID {
	return network.MSG_HEAD
}

// Parse parses the packet from binary
func (m *MsgHead) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgHead) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
