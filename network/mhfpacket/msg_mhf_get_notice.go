package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetNotice represents the MSG_MHF_GET_NOTICE
type MsgMhfGetNotice struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetNotice) Opcode() network.PacketID {
	return network.MSG_MHF_GET_NOTICE
}

// Parse parses the packet from binary
func (m *MsgMhfGetNotice) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetNotice) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
