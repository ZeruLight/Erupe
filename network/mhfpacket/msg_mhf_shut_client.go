package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfShutClient represents the MSG_MHF_SHUT_CLIENT
type MsgMhfShutClient struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfShutClient) Opcode() network.PacketID {
	return network.MSG_MHF_SHUT_CLIENT
}

// Parse parses the packet from binary
func (m *MsgMhfShutClient) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfShutClient) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
