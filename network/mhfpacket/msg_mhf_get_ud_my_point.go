package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdMyPoint represents the MSG_MHF_GET_UD_MY_POINT
type MsgMhfGetUdMyPoint struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdMyPoint) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_MY_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdMyPoint) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdMyPoint) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}