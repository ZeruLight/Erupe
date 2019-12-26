package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysTime represents the MSG_SYS_TIME
type MsgSysTime struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysTime) Opcode() network.PacketID {
	return network.MSG_SYS_TIME
}

// Parse parses the packet from binary
func (m *MsgSysTime) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysTime) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}