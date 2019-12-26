package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysEnumerateClient represents the MSG_SYS_ENUMERATE_CLIENT
type MsgSysEnumerateClient struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysEnumerateClient) Opcode() network.PacketID {
	return network.MSG_SYS_ENUMERATE_CLIENT
}

// Parse parses the packet from binary
func (m *MsgSysEnumerateClient) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysEnumerateClient) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}