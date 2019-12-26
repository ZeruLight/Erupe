package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysPing represents the MSG_SYS_PING
type MsgSysPing struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysPing) Opcode() network.PacketID {
	return network.MSG_SYS_PING
}

// Parse parses the packet from binary
func (m *MsgSysPing) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysPing) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}