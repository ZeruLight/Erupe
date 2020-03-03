package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReleaseEvent represents the MSG_MHF_RELEASE_EVENT
type MsgMhfReleaseEvent struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReleaseEvent) Opcode() network.PacketID {
	return network.MSG_MHF_RELEASE_EVENT
}

// Parse parses the packet from binary
func (m *MsgMhfReleaseEvent) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReleaseEvent) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
