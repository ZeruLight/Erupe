package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateGuacot represents the MSG_MHF_ENUMERATE_GUACOT
type MsgMhfEnumerateGuacot struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuacot) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUACOT
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuacot) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuacot) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}