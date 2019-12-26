package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateOrder represents the MSG_MHF_ENUMERATE_ORDER
type MsgMhfEnumerateOrder struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateOrder) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_ORDER
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateOrder) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateOrder) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}