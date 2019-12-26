package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateEtcPoint represents the MSG_MHF_UPDATE_ETC_POINT
type MsgMhfUpdateEtcPoint struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateEtcPoint) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_ETC_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateEtcPoint) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateEtcPoint) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}