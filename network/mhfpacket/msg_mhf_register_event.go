package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfRegisterEvent represents the MSG_MHF_REGISTER_EVENT
type MsgMhfRegisterEvent struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfRegisterEvent) Opcode() network.PacketID {
	return network.MSG_MHF_REGISTER_EVENT
}

// Parse parses the packet from binary
func (m *MsgMhfRegisterEvent) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfRegisterEvent) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}