package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetExtraInfo represents the MSG_MHF_GET_EXTRA_INFO
type MsgMhfGetExtraInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetExtraInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_EXTRA_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetExtraInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetExtraInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
