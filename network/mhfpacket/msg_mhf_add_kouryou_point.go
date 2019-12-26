package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAddKouryouPoint represents the MSG_MHF_ADD_KOURYOU_POINT
type MsgMhfAddKouryouPoint struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddKouryouPoint) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_KOURYOU_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfAddKouryouPoint) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddKouryouPoint) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}