package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAddUdPoint represents the MSG_MHF_ADD_UD_POINT
type MsgMhfAddUdPoint struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddUdPoint) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_UD_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfAddUdPoint) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddUdPoint) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
