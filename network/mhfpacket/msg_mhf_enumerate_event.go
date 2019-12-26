package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateEvent represents the MSG_MHF_ENUMERATE_EVENT
type MsgMhfEnumerateEvent struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateEvent) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_EVENT
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateEvent) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateEvent) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}